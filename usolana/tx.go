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
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/memo"
	"github.com/gagliardetto/solana-go/programs/stake"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/programs/tokenregistry"
	"github.com/gagliardetto/solana-go/programs/vote"
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
	PriorityFee          uint64 // transaction priority fee // = (compute unit price * compute unit limit) / microLamportsPerLamport // lamports
	ComputeUnitsConsumed uint64 // compute units consumed
	TxVersion            int    // transaction version // -1: legacy
}

type TxParser struct {
	cli              *rpc.Client
	insParserFactory instructionsParserFactory
}

func NewTxParser(endpoint string) *TxParser {
	return &TxParser{
		cli:              rpc.New(endpoint),
		insParserFactory: newInstructionsParserFactory(),
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

		if programID == computebudget.ProgramID.String() {
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
	return tp.insParserFactory.GetParser(programID)(tx, ins)
}

func parseInstructionAccounts(accs []*solana.AccountMeta) []ParsedInstructionAccount {
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

type instructionsParser func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error)

type instructionsParserFactory struct {
	parsers map[string]instructionsParser
}

func (f *instructionsParserFactory) GetParser(programID string) instructionsParser {
	parser, ok := f.parsers[programID]
	if !ok {
		return func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
			return ParsedInstruction{
				ProgramID: programID,
				Name:      "Unknown",
				Data:      ins.Data.String(), // base58
			}, nil
		}
	}
	return parser
}

func newInstructionsParserFactory() instructionsParserFactory {
	return instructionsParserFactory{
		parsers: map[string]instructionsParser{
			system.ProgramID.String(): func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
				var insData system.Instruction
				err := insData.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
				if err != nil {
					return ParsedInstruction{}, fmt.Errorf("unmarshal system instruction: %w", err)
				}
				insID := insData.TypeID.Uint32()
				return ParsedInstruction{
					ProgramID: insData.ProgramID().String(),
					TypeID:    insID,
					Name:      system.InstructionIDToName(insID),
					Accounts:  parseInstructionAccounts(insData.Accounts()),
					Data:      insData.Impl,
				}, nil
			},
			token.ProgramID.String(): func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
				var insData token.Instruction
				err := insData.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
				if err != nil {
					return ParsedInstruction{}, fmt.Errorf("unmarshal token instruction: %w", err)
				}
				insID := insData.TypeID.Uint8()
				return ParsedInstruction{
					ProgramID: insData.ProgramID().String(),
					TypeID:    uint32(insID),
					Name:      token.InstructionIDToName(insID),
					Accounts:  parseInstructionAccounts(insData.Accounts()),
					Data:      insData.Impl,
				}, nil
			},
			computebudget.ProgramID.String(): func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
				var insData computebudget.Instruction
				err := insData.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
				if err != nil {
					return ParsedInstruction{}, fmt.Errorf("unmarshal compute budget instruction: %w", err)
				}
				insID := insData.TypeID.Uint8()
				return ParsedInstruction{
					ProgramID: insData.ProgramID().String(),
					TypeID:    uint32(insID),
					Name:      computebudget.InstructionIDToName(insID),
					Accounts:  parseInstructionAccounts(insData.Accounts()),
					Data:      insData.Impl,
				}, nil
			},
			associatedtokenaccount.ProgramID.String(): func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
				var insData associatedtokenaccount.Instruction
				err := insData.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
				if err != nil {
					return ParsedInstruction{}, fmt.Errorf("unmarshal associated token account instruction: %w", err)
				}
				insID := insData.TypeID.Uint8()
				bv := new(bin.BaseVariant)
				bv.Assign(insData.TypeID, nil)
				_, name, _ := bv.Obtain(associatedtokenaccount.InstructionImplDef)
				return ParsedInstruction{
					ProgramID: insData.ProgramID().String(),
					TypeID:    uint32(insID),
					Name:      name,
					Accounts:  parseInstructionAccounts(insData.Accounts()),
					Data:      insData.Impl,
				}, nil
			},
			memo.ProgramID.String(): func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
				var insData memo.MemoInstruction
				err := insData.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
				if err != nil {
					return ParsedInstruction{}, fmt.Errorf("unmarshal memo instruction: %w", err)
				}
				insID := insData.TypeID.Uint8()
				bv := new(bin.BaseVariant)
				bv.Assign(insData.TypeID, nil)
				_, name, _ := bv.Obtain(memo.InstructionImplDef)
				return ParsedInstruction{
					ProgramID: insData.ProgramID().String(),
					TypeID:    uint32(insID),
					Name:      name,
					Accounts:  parseInstructionAccounts(insData.Accounts()),
					Data:      insData.Impl,
				}, nil
			},
			stake.ProgramID.String(): func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
				var insData stake.Instruction
				err := insData.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
				if err != nil {
					return ParsedInstruction{}, fmt.Errorf("unmarshal stake instruction: %w", err)
				}
				insID := insData.TypeID.Uint32()
				bv := new(bin.BaseVariant)
				bv.Assign(insData.TypeID, nil)
				_, name, _ := bv.Obtain(stake.InstructionImplDef)
				return ParsedInstruction{
					ProgramID: insData.ProgramID().String(),
					TypeID:    insID,
					Name:      name,
					Accounts:  parseInstructionAccounts(insData.Accounts()),
					Data:      insData.Impl,
				}, nil
			},
			vote.ProgramID.String(): func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
				var insData vote.Instruction
				err := insData.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
				if err != nil {
					return ParsedInstruction{}, fmt.Errorf("unmarshal vote instruction: %w", err)
				}
				insID := insData.TypeID.Uint32()
				bv := new(bin.BaseVariant)
				bv.Assign(insData.TypeID, nil)
				_, name, _ := bv.Obtain(vote.InstructionImplDef)
				return ParsedInstruction{
					ProgramID: insData.ProgramID().String(),
					TypeID:    insID,
					Name:      name,
					Accounts:  parseInstructionAccounts(insData.Accounts()),
					Data:      insData.Impl,
				}, nil
			},
			// token registry program id (mainnet beta) // NOTE: see tokenregistry.ProgramID()
			"CmPVzy88JSB4S223yCvFmBxTLobLya27KgEDeNPnqEub": func(tx *solana.Transaction, ins solana.CompiledInstruction) (ParsedInstruction, error) {
				var insData tokenregistry.Instruction
				err := insData.UnmarshalWithDecoder(bin.NewBinDecoder(ins.Data))
				if err != nil {
					return ParsedInstruction{}, fmt.Errorf("unmarshal token registry instruction: %w", err)
				}
				insID := insData.TypeID.Uint32()
				bv := new(bin.BaseVariant)
				bv.Assign(insData.TypeID, nil)
				_, name, _ := bv.Obtain(tokenregistry.InstructionDefVariant)
				return ParsedInstruction{
					ProgramID: insData.ProgramID().String(),
					TypeID:    insID,
					Name:      name,
					Accounts:  parseInstructionAccounts(insData.Accounts()),
					Data:      insData.Impl,
				}, nil
			},
		},
	}
}
