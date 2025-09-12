package uethereum

import (
	"context"
	"fmt"
	"math/big"
	"slices"

	"github.com/15ho/wallet-utils-go/internal/zlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"
)

type ParsedLog struct {
	Topics []string // log topics // hex string
	Data   string   // log data // hex string
}

type ParsedTx struct {
	Block          *big.Int     // block height
	Timestamp      int64        // block timestamp // milliseconds
	TxHash         string       // transaction hash
	From           string       // from address
	To             string       // to address // wallet or contract address
	Value          *big.Int     // transaction value
	Fee            *big.Int     // transaction fee // = gas price * gas used
	GasLimit       uint64       // transaction gas limit
	GasUsed        uint64       // transaction gas used
	GasPrice       *big.Int     // transaction gas price // static or dynamic // dynamic: min((base fee + max priority fee), max fee)
	BaseFee        *big.Int     // transaction base fee
	MaxPriorityFee *big.Int     // transaction max priority fee
	MaxFee         *big.Int     // transaction max fee
	InputData      string       // transaction input data // hex string
	Logs           []*ParsedLog // transaction logs
	Nonce          uint64
	TxType         uint8 // transaction type // https://ethereum.org/developers/docs/transactions/#typed-transaction-envelope
}

type TxParser struct {
	cli *ethclient.Client
}

func NewTxParser(endpoint string) (*TxParser, error) {
	cli, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	return &TxParser{cli}, nil
}

func (tp *TxParser) ParseBlock(ctx context.Context, blockNumber *big.Int) ([]*ParsedTx, error) {
	block, err := tp.cli.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("get block by number(%v): %w", blockNumber, err)
	}

	bn := rpc.LatestBlockNumber
	if blockNumber != nil {
		bn = rpc.BlockNumber(blockNumber.Int64())
	}

	receipts, err := tp.cli.BlockReceipts(ctx, rpc.BlockNumberOrHash{
		BlockNumber: &bn,
	})
	if err != nil {
		return nil, fmt.Errorf("get block receipts  by number(%s): %w", bn.String(), err)
	}

	parsedTxs := make([]*ParsedTx, 0, len(block.Transactions()))
	slices.All(block.Transactions())(func(idx int, tx *types.Transaction) bool {
		ptx, err := tp.parseTx(block.Header(), tx, receipts[idx])
		if err != nil {
			zlog.Error("parseTxWithReceipt", zap.Error(err), zap.String("txHash", tx.Hash().Hex()))
			return true
		}
		parsedTxs = append(parsedTxs, ptx)
		return true
	})
	return parsedTxs, nil
}

func (tp *TxParser) parseTx(blockHeader *types.Header, tx *types.Transaction, receipt *types.Receipt) (ptx *ParsedTx, err error) {
	if tx.Hash().Cmp(receipt.TxHash) != 0 {
		err = fmt.Errorf("tx hash(%s) is not equal receipt's tx hash(%s)", tx.Hash().Hex(), receipt.TxHash.Hex())
		return
	}

	from, err := getTxFrom(tx)
	if err != nil {
		err = fmt.Errorf("getTxSender error: %w", err)
		return
	}

	logs := make([]*ParsedLog, 0, len(receipt.Logs))
	slices.Values(receipt.Logs)(func(log *types.Log) bool {
		topics := make([]string, 0, len(log.Topics))
		slices.Values(log.Topics)(func(topic common.Hash) bool {
			topics = append(topics, topic.Hex())
			return true
		})
		logs = append(logs, &ParsedLog{
			Data:   hexutil.Encode(log.Data),
			Topics: topics,
		})
		return true
	})

	ptx = &ParsedTx{
		Block:          blockHeader.Number,
		Timestamp:      int64(blockHeader.Time),
		TxHash:         tx.Hash().Hex(),
		From:           from.Hex(),
		To:             tx.To().Hex(),
		Value:          tx.Value(),
		GasLimit:       tx.Gas(),
		GasUsed:        receipt.GasUsed,
		GasPrice:       receipt.EffectiveGasPrice,
		BaseFee:        blockHeader.BaseFee,
		MaxPriorityFee: tx.GasTipCap(),
		MaxFee:         tx.GasFeeCap(),
		InputData:      hexutil.Encode(tx.Data()),
		Logs:           logs,
		Nonce:          tx.Nonce(),
		TxType:         tx.Type(),
	}

	ptx.Fee = new(big.Int).Mul(ptx.GasPrice, new(big.Int).SetUint64(ptx.GasUsed))

	return
}

func getTxFrom(tx *types.Transaction) (from common.Address, err error) {
	from, err = types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		chainID := tx.ChainId()
		if chainID == nil || chainID.Sign() <= 0 {
			return
		}
		from, err = types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
		if err != nil {
			return
		}
	}
	return
}
