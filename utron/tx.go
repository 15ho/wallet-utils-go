package utron

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"

	"github.com/15ho/wallet-utils-go/internal/zlog"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ParsedLog struct {
	Topics []string // log topics // hex string
	Data   string   // log data // hex string
}

type ParsedTxFee struct {
	Fee           int64 // transaction fee
	EnergyUsed    int64 // energy used
	BandwidthUsed int64 // bandwidth used
}

type ParsedTx struct {
	Block     int64        // block height
	Timestamp int64        // block timestamp // milliseconds
	TxHash    string       // transaction hash
	Status    string       // transaction status // success or fail
	From      string       // from address
	To        string       // to address // wallet or contract address
	Value     int64        // transaction value
	Fee       ParsedTxFee  // transaction fee
	FeeLimit  int64        // transaction fee limit
	InputData string       // transaction input data // hex string
	Logs      []*ParsedLog // transaction logs
	TxType    int32        // transaction type // https://github.com/tronprotocol/java-tron/blob/develop/protocol/src/main/protos/core/Tron.proto#L338
}

type TxParser struct {
	cli *client.GrpcClient
}

func newTxParser(endpoint string, opts ...grpc.DialOption) (tp *TxParser, cleanup func(), err error) {
	cli := client.NewGrpcClient(endpoint)
	if err = cli.Start(opts...); err != nil {
		return
	}
	cleanup = func() {
		cli.Stop()
	}
	tp = &TxParser{
		cli: cli,
	}
	return
}

func NewTxParser(endpoint string) (tp *TxParser, cleanup func(), err error) {
	return newTxParser(endpoint, client.GRPCInsecure())
}

func NewTxParserWithBasicAuth(endpoint, token string) (tp *TxParser, cleanup func(), err error) {
	return newTxParser(endpoint, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(basicAuth{
			username: endpoint,
			password: token,
		}))
}

func NewTxParserWithXToken(endpoint, token string) (tp *TxParser, cleanup func(), err error) {
	return newTxParser(endpoint, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(auth{token}))
}

func (tp *TxParser) ParseBlock(ctx context.Context) ([]*ParsedTx, error) {
	block, err := tp.cli.GetNowBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	blockInfo, err := tp.cli.GetBlockInfoByNum(block.BlockHeader.RawData.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to get block info: %w", err)
	}

	parsedTxs := make([]*ParsedTx, 0, len(block.Transactions))
	slices.All(block.Transactions)(func(idx int, txExt *api.TransactionExtention) bool {
		ptx, err := tp.parseTx(txExt.Transaction, blockInfo.TransactionInfo[idx])
		if err != nil {
			zlog.Error("parseTxWithTxInfo", zap.Error(err), zap.String("txHash", hex.EncodeToString(txExt.Txid)))
			return true
		}
		parsedTxs = append(parsedTxs, ptx)
		return true
	})

	return parsedTxs, nil
}

func (tp *TxParser) parseTx(tx *core.Transaction, txInfo *core.TransactionInfo) (ptx *ParsedTx, err error) {
	txContracts := tx.RawData.Contract
	if len(txContracts) == 0 {
		err = errors.New("transaction contract is empty")
		return
	}
	txStatus := "fail"
	if len(tx.Ret) > 0 &&
		tx.Ret[0].ContractRet == core.Transaction_Result_SUCCESS {
		txStatus = "success"
	}

	parsedLogs := make([]*ParsedLog, 0, len(txInfo.Log))
	slices.Values(txInfo.Log)(func(txLog *core.TransactionInfo_Log) bool {
		topics := make([]string, 0, len(txLog.Topics))
		slices.Values(txLog.Topics)(func(topic []byte) bool {
			topics = append(topics, hex.EncodeToString(topic))
			return true
		})
		parsedLogs = append(parsedLogs, &ParsedLog{
			Topics: topics,
			Data:   hex.EncodeToString(txLog.Data),
		})
		return true
	})

	ptx = &ParsedTx{
		Block:     txInfo.BlockNumber,
		Timestamp: txInfo.BlockTimeStamp,
		TxHash:    hex.EncodeToString(txInfo.Id),
		Status:    txStatus,
		Fee: ParsedTxFee{
			Fee:           txInfo.Fee,
			EnergyUsed:    txInfo.Receipt.EnergyUsageTotal,
			BandwidthUsed: txInfo.Receipt.NetUsage,
		},
		FeeLimit:  tx.RawData.FeeLimit,
		InputData: hex.EncodeToString(txContracts[0].GetParameter().Value),
		Logs:      parsedLogs,
		TxType:    int32(txContracts[0].Type),
	}

	switch txContracts[0].Type {
	case core.Transaction_Contract_TriggerSmartContract:
		var triggerSmartContract core.TriggerSmartContract
		err = txContracts[0].GetParameter().UnmarshalTo(&triggerSmartContract)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal transfer contract: %w", err)
			return
		}
		ptx.From = address.Address(triggerSmartContract.OwnerAddress).String()
		ptx.To = address.Address(triggerSmartContract.ContractAddress).String()
	case core.Transaction_Contract_TransferContract:
		var transfer core.TransferContract
		err = txContracts[0].GetParameter().UnmarshalTo(&transfer)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal transfer contract: %w", err)
			return
		}
		ptx.From = address.Address(transfer.OwnerAddress).String()
		ptx.To = address.Address(transfer.ToAddress).String()
		ptx.Value = transfer.Amount
	case core.Transaction_Contract_DelegateResourceContract:
		var delegateRes core.DelegateResourceContract
		err = txContracts[0].GetParameter().UnmarshalTo(&delegateRes)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal transfer contract: %w", err)
			return
		}
		ptx.From = address.Address(delegateRes.OwnerAddress).String()
		ptx.To = address.Address(delegateRes.ReceiverAddress).String()
	case core.Transaction_Contract_UnDelegateResourceContract:
		var unDelegateRes core.UnDelegateResourceContract
		err = txContracts[0].GetParameter().UnmarshalTo(&unDelegateRes)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal transfer contract: %w", err)
			return
		}
		ptx.From = address.Address(unDelegateRes.OwnerAddress).String()
		ptx.To = address.Address(unDelegateRes.ReceiverAddress).String()
	default:
		// TODO: support other transaction types
	}

	return
}
