package usolana

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/15ho/wallet-utils-go/internal/zlog"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/zap"
)

// TxPriorityFee Transaction Priority Fee
// The prioritization fee is an optional fee paid to increase the chance that the current leader processes your transaction.
// https://solana.com/docs/core/fees#prioritization-fee
type TxPriorityFee struct {
	ComputeUnitLimit uint32
	ComputeUnitPrice uint64 // microLamports // 1,000,000 micro-lamports = 1 lamport
}

type ParsedInstructionAccount struct {
	Address    string // base58
	IsWritable bool
	IsSigner   bool
}

type ParsedInstruction struct {
	ProgramID         string
	TypeID            uint32                     // instruction type id
	Name              string                     // instruction name
	Accounts          []ParsedInstructionAccount // instruction input accounts
	Data              any                        // instruction input data
	InnerInstructions []ParsedInstruction
}

// ParsedTx parsed transaction
type ParsedTx struct {
	Block                uint64   // block height // slot
	Timestamp            int64    // block timestamp // milliseconds
	TxHash               string   // transaction hash
	Status               string   // transaction status // success or fail
	Signer               []string // signer addresses
	Instructions         []ParsedInstruction
	Fee                  uint64 // transaction fee // = base fee + priority fee // lamports
	PriorityFee          uint64 // transaction priority fee // lamports
	ComputeUnitsConsumed uint64 // compute units consumed
	TxVersion            int    // transaction version // -1: legacy
}

type TxParser struct {
	cli *rpc.Client
}

func NewTxParser(endpoint string) *TxParser {
	return &TxParser{
		cli: rpc.New(endpoint),
	}
}

func NewTestnetTxParser() *TxParser {
	return NewTxParser(rpc.TestNet_RPC)
}

func NewDevnetTxParser() *TxParser {
	return NewTxParser(rpc.DevNet_RPC)
}

func (tp *TxParser) ParseLatestBlock(ctx context.Context) ([]*ParsedTx, error) {
	slot, err := tp.cli.GetSlot(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, err
	}
	return tp.ParseBlock(ctx, slot)
}

func (tp *TxParser) ParseBlock(ctx context.Context, slot uint64) ([]*ParsedTx, error) {
	res, err := tp.cli.GetBlockWithOpts(ctx, slot, &rpc.GetBlockOpts{
		MaxSupportedTransactionVersion: rpc.NewTransactionVersion(rpc.MaxSupportedTransactionVersion0),
		Rewards:                        rpc.NewBoolean(false),
		Commitment:                     rpc.CommitmentFinalized, // default
	})
	if err != nil {
		return nil, err
	}
	parsedTxs := make([]*ParsedTx, 0, len(res.Transactions))
	slices.Values(res.Transactions)(func(twm rpc.TransactionWithMeta) bool {
		parsedTx, err := tp.parseConfirmedTx(twm)
		if err != nil {
			zlog.Error("parseConfirmedTx error", zap.Error(err))
			return true
		}
		parsedTxs = append(parsedTxs, parsedTx)
		return true
	})
	return parsedTxs, nil
}

func (tp *TxParser) parseConfirmedTx(twm rpc.TransactionWithMeta) (ptx *ParsedTx, err error) {
	if twm.Meta == nil {
		err = errors.New("meta is nil")
		return
	}

	tx, err := twm.GetTransaction()
	if err != nil {
		err = fmt.Errorf("get transaction: %w", err)
		return
	}
	if len(tx.Signatures) == 0 {
		err = errors.New("signatures is empty")
		return
	}

	parsedInss := make([]ParsedInstruction, 0, len(tx.Message.Instructions))

	lookupTableInners := make(map[uint16]rpc.InnerInstruction, len(twm.Meta.InnerInstructions))
	for _, inner := range twm.Meta.InnerInstructions {
		lookupTableInners[inner.Index] = inner
	}

	var (
		computeUnitPrice uint64
		computeUnitLimit uint32
	)

	slices.All(tx.Message.Instructions)(func(idx int, ins solana.CompiledInstruction) (next bool) {
		next = true

		programID := tx.Message.AccountKeys[ins.ProgramIDIndex].String()

		zlog.Debug("parse instruction",
			zap.String("programID", programID),
			zap.Uint16("programIDIndex", ins.ProgramIDIndex))

		parsedIns, err := tp.parseInstruction(tx, ins)
		if err != nil {
			zlog.Error("parse instruction error",
				zap.Error(err),
				zap.String("programID", programID),
				zap.Uint16("programIDIndex", ins.ProgramIDIndex))
			return
		}

		if programID == ProgramIDComputeBudget {
			switch uint8(parsedIns.TypeID) {
			case computebudget.Instruction_SetComputeUnitLimit:
				computeUnitLimit = parsedIns.Data.(*computebudget.SetComputeUnitLimit).Units
			case computebudget.Instruction_SetComputeUnitPrice:
				computeUnitPrice = parsedIns.Data.(*computebudget.SetComputeUnitPrice).MicroLamports
			}
		}

		innerIns, ok := lookupTableInners[uint16(idx)]
		if ok && len(innerIns.Instructions) > 0 {
			zlog.Debug("found inner instructions",
				zap.String("programID", programID),
				zap.Int("index", idx),
				zap.Int("count", len(innerIns.Instructions)))
			parsedInnerInss := make([]ParsedInstruction, 0, len(innerIns.Instructions))
			for _, innerIns := range innerIns.Instructions {
				innerProgramID := tx.Message.AccountKeys[innerIns.ProgramIDIndex].String()
				parsedInnerIns, err := tp.parseInstruction(tx, solana.CompiledInstruction{
					ProgramIDIndex: innerIns.ProgramIDIndex,
					Accounts:       innerIns.Accounts,
					Data:           innerIns.Data,
				})
				if err != nil {
					zlog.Error("parse inner instruction error",
						zap.Error(err),
						zap.String("programID", innerProgramID),
						zap.Uint16("programIDIndex", innerIns.ProgramIDIndex))
					continue
				}
				parsedInnerInss = append(parsedInnerInss, parsedInnerIns)
			}
			parsedIns.InnerInstructions = parsedInnerInss
		}

		parsedInss = append(parsedInss, parsedIns)
		return
	})

	txStatus := "success"
	if twm.Meta.Err != nil {
		txStatus = "fail"
	}

	priorityFeeMicroLamports := new(big.Int).Mul(new(big.Int).SetUint64(computeUnitPrice), new(big.Int).SetUint64(uint64(computeUnitLimit)))
	priorityFee := new(big.Int).Div(priorityFeeMicroLamports, new(big.Int).SetUint64(MicroLamportsPerLamport)).Uint64()

	ptx = &ParsedTx{
		Block:                twm.Slot,
		Fee:                  twm.Meta.Fee,
		Timestamp:            twm.BlockTime.Time().UnixMilli(),
		TxHash:               tx.Signatures[0].String(),
		Status:               txStatus,
		Signer:               tx.Message.Signers().ToBase58(),
		Instructions:         parsedInss,
		PriorityFee:          priorityFee,
		ComputeUnitsConsumed: *twm.Meta.ComputeUnitsConsumed,
		TxVersion:            int(twm.Version),
	}
	return
}

func (tp *TxParser) parseInstruction(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
	programID := tx.Message.AccountKeys[ins.ProgramIDIndex].String()
	switch programID {
	case ProgramIDSystem:
		var sysIns system.Instruction
		err := sysIns.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
		if err != nil {
			return ParsedInstruction{}, fmt.Errorf("unmarshal system instruction: %w", err)
		}
		insID := sysIns.TypeID.Uint32()
		return ParsedInstruction{
			ProgramID: programID,
			TypeID:    insID,
			Name:      system.InstructionIDToName(insID),
			Accounts:  tp.parseInstructionAccounts(sysIns.Accounts()),
			Data:      sysIns.Impl,
		}, nil
	case ProgramIDToken:
		var tokenIns token.Instruction
		err := tokenIns.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
		if err != nil {
			return ParsedInstruction{}, fmt.Errorf("unmarshal system instruction: %w", err)
		}
		insID := tokenIns.TypeID.Uint8()
		return ParsedInstruction{
			ProgramID: programID,
			TypeID:    uint32(insID),
			Name:      token.InstructionIDToName(insID),
			Accounts:  tp.parseInstructionAccounts(tokenIns.Accounts()),
			Data:      tokenIns.Impl,
		}, nil
	case ProgramIDComputeBudget:
		var cbIns computebudget.Instruction
		err := cbIns.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
		if err != nil {
			return ParsedInstruction{}, fmt.Errorf("unmarshal system instruction: %w", err)
		}
		insID := cbIns.TypeID.Uint8()
		return ParsedInstruction{
			ProgramID: programID,
			TypeID:    uint32(insID),
			Name:      computebudget.InstructionIDToName(insID),
			Accounts:  tp.parseInstructionAccounts(cbIns.Accounts()),
			Data:      cbIns.Impl,
		}, nil
	default:
		// TODO: implement other programs
		return ParsedInstruction{
			ProgramID: programID,
			Name:      "Unknown",
			Data:      ins.Data.String(), // base58
		}, nil
	}
}

func (*TxParser) parseInstructionAccounts(accs []*solana.AccountMeta) []ParsedInstructionAccount {
	if len(accs) == 0 {
		return nil
	}
	parsedAccs := make([]ParsedInstructionAccount, 0, len(accs))
	slices.Values(accs)(func(acc *solana.AccountMeta) bool {
		parsedAccs = append(parsedAccs, ParsedInstructionAccount{
			Address:    acc.PublicKey.String(),
			IsSigner:   acc.IsSigner,
			IsWritable: acc.IsWritable,
		})
		return true
	})
	return parsedAccs
}
