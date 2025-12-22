package utron

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	tronaddr "github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

func CreateWalletAccount() (privateKeyHex, address string, err error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return
	}
	pkBytes := crypto.FromECDSA(pk)
	privateKeyHex = hexutil.Encode(pkBytes)

	// base58(bytes(0x41) + common.Address.bytes())
	address = tronaddr.PubkeyToAddress(pk.PublicKey).String()
	return
}

func WalletAddressFromPrivateKey(privateKeyHex string) (address string, err error) {
	pk, err := crypto.ToECDSA(common.FromHex(privateKeyHex))
	if err != nil {
		return
	}
	address = tronaddr.PubkeyToAddress(pk.PublicKey).String()
	return
}

type WalletClient struct {
	cli *client.GrpcClient

	privateKey *ecdsa.PrivateKey
	account    string
}

func newWalletClient(endpoint string, privateKeyHex string, opts ...grpc.DialOption) (wc *WalletClient, cleanup func(), err error) {
	pk, err := crypto.ToECDSA(common.FromHex(privateKeyHex))
	if err != nil {
		return
	}
	cli := client.NewGrpcClient(endpoint)
	if err = cli.Start(opts...); err != nil {
		return
	}
	cleanup = func() {
		cli.Stop()
	}
	wc = &WalletClient{
		cli:        cli,
		privateKey: pk,
		account:    tronaddr.PubkeyToAddress(pk.PublicKey).String(),
	}
	return
}

func NewWalletClient(endpoint, privateKeyHex string) (wc *WalletClient, cleanup func(), err error) {
	return newWalletClient(endpoint, privateKeyHex, client.GRPCInsecure())
}

func NewWalletClientWithBasicAuth(endpoint, token, privateKeyHex string) (wc *WalletClient, cleanup func(), err error) {
	return newWalletClient(endpoint, privateKeyHex, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(basicAuth{
			username: endpoint,
			password: token,
		}))
}

func NewWalletClientWithXToken(endpoint, token, privateKeyHex string) (wc *WalletClient, cleanup func(), err error) {
	return newWalletClient(endpoint, privateKeyHex, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(auth{token}))
}

type Gas struct {
	Total int64 // = Bandwidth * BandwidthUnitPrice + Energy + EnergyUnitPrice
	// or, = CreateAccountFee + CreateAccountBandwidthFee
	// or, = CreateAccountFee + (Bandwidth * BandwidthUnitPrice)

	BandwidthUsed      int64
	BandwidthUnitPrice int64

	// --- contract deploy or call ---
	EnergyUsed      int64
	EnergyUnitPrice int64

	// --- only transfer trx to inactivated account ---
	CreateAccountFee          int64
	CreateAccountBandwidthFee int64
}

func (wc *WalletClient) EstimateGasTransferTRX(ctx context.Context, to string, amount int64) (gas Gas, err error) {
	params, err := wc.cli.Client.GetChainParameters(ctx, &api.EmptyMessage{})
	if err != nil {
		err = fmt.Errorf("get chain parameters error: %w", err)
		return
	}

	_, checkErr := wc.cli.GetAccount(to)
	if checkErr != nil {
		if !strings.Contains("account not found", checkErr.Error()) {
			err = checkErr
			return
		}
		// --- inactivated account ---
		var (
			createAccountFee          int64 // getCreateNewAccountFeeInSystemContract
			createAccountBandwidthFee int64 // getCreateAccountFee
			bandwidthUnitPrice        int64 // getTransactionFee
		)
		// https://developers.tron.network/reference/wallet-getchainparameters
		slices.Values(params.GetChainParameter())(func(param *core.ChainParameters_ChainParameter) bool {
			if param != nil {
				switch param.Key {
				case "getCreateNewAccountFeeInSystemContract":
					createAccountFee = param.Value
				case "getCreateAccountFee":
					createAccountBandwidthFee = param.Value
				case "getTransactionFee":
					bandwidthUnitPrice = param.Value
				}
				return createAccountFee == 0 ||
					createAccountBandwidthFee == 0 ||
					bandwidthUnitPrice == 0
			}
			return true
		})

		txExt, terr := wc.cli.Transfer(wc.account, to, amount)
		if terr != nil {
			err = fmt.Errorf("create transfer tx error: %w", terr)
			return
		}
		signature, terr := crypto.Sign(txExt.Txid, wc.privateKey)
		if terr != nil {
			err = fmt.Errorf("sign error: %w", terr)
			return
		}
		tx := txExt.Transaction
		tx.Signature = append(tx.Signature, signature)
		txBytes, terr := proto.Marshal(tx)
		if terr != nil {
			err = fmt.Errorf("marshal tx error: %w", terr)
			return
		}
		bandwidthUsed := int64(len(txBytes) + 64)

		accRes, terr := wc.cli.GetAccountResource(wc.account)
		if terr != nil {
			err = fmt.Errorf("getAccountResource failed: %w", terr)
			return
		}

		// When creating an account, if the creator’s account does not have enough bandwidth obtained through staking,
		// then the value corresponding to this parameter needs to be burned to pay for the bandwidth
		stakeBandwidth := accRes.NetLimit - accRes.NetUsed
		if stakeBandwidth < bandwidthUsed {
			gas = Gas{
				Total:                     createAccountFee + createAccountBandwidthFee,
				CreateAccountFee:          createAccountFee,
				CreateAccountBandwidthFee: createAccountBandwidthFee,
			}
			return
		}

		gas = Gas{
			Total:              createAccountFee + bandwidthUsed*bandwidthUnitPrice,
			CreateAccountFee:   createAccountFee,
			BandwidthUsed:      bandwidthUsed,
			BandwidthUnitPrice: bandwidthUnitPrice,
		}
		return
	}

	// --- activated account ---
	var bandwidthUnitPrice int64 // getTransactionFee
	// https://developers.tron.network/reference/wallet-getchainparameters
	slices.Values(params.GetChainParameter())(func(param *core.ChainParameters_ChainParameter) bool {
		if param != nil && param.Key == "getTransactionFee" {
			bandwidthUnitPrice = param.Value
			return false
		}
		return true
	})

	// bandwidth fee
	// https://developers.tron.network/docs/tron-protocol-transaction#bandwidth-fee
	txExt, err := wc.cli.Transfer(wc.account, to, amount)
	if err != nil {
		err = fmt.Errorf("create transfer tx error: %w", err)
		return
	}
	signature, err := crypto.Sign(txExt.Txid, wc.privateKey)
	if err != nil {
		err = fmt.Errorf("sign error: %w", err)
		return
	}
	tx := txExt.Transaction
	tx.Signature = append(tx.Signature, signature)
	txBytes, err := proto.Marshal(tx)
	if err != nil {
		err = fmt.Errorf("marshal tx error: %w", err)
		return
	}
	// TRX burned to pay for Bandwidth = Total Bandwidth consumed by the transaction * Bandwidth unit price
	// https://developers.tron.network/docs/faq#5-how-to-calculate-the-bandwidth-and-energy-consumed-when-callingdeploying-a-contract
	bandwidthUsed := int64(len(txBytes) + 64)
	gas = Gas{
		Total:              bandwidthUsed * bandwidthUnitPrice,
		BandwidthUsed:      bandwidthUsed,
		BandwidthUnitPrice: bandwidthUnitPrice,
	}
	return
}

func (wc *WalletClient) TransferTRX(ctx context.Context, to string, amount int64) (txHash string, err error) {
	txExt, err := wc.cli.Transfer(wc.account, to, amount)
	if err != nil {
		err = fmt.Errorf("create transfer tx error: %w", err)
		return
	}
	signature, err := crypto.Sign(txExt.Txid, wc.privateKey)
	if err != nil {
		err = fmt.Errorf("sign error: %w", err)
		return
	}
	txExt.Transaction.Signature = append(txExt.Transaction.Signature, signature)
	ret, err := wc.cli.Broadcast(txExt.GetTransaction())
	if err != nil {
		err = fmt.Errorf("broadcast trx error: %v", err)
		return
	}
	if !ret.Result {
		err = fmt.Errorf("broadcast trx fail: %s", ret.String())
		return
	}
	txHash = hex.EncodeToString(txExt.GetTxid())
	return
}

func (wc *WalletClient) GetTRXBalance(ctx context.Context) (balance int64, err error) {
	acc, err := wc.cli.GetAccount(wc.account)
	if err != nil {
		return
	}
	balance = acc.Balance
	return
}

func (wc *WalletClient) GetTRXBalanceByAddress(ctx context.Context, address string) (balance int64, err error) {
	acc, err := wc.cli.GetAccount(address)
	if err != nil {
		return
	}
	balance = acc.Balance
	return
}

func (wc *WalletClient) EstimateGasTransferTRC20Token(ctx context.Context, tokenAddress, to string, amount *big.Int, feeLimit int64) (gas Gas, err error) {
	params, err := wc.cli.Client.GetChainParameters(ctx, &api.EmptyMessage{})
	if err != nil {
		err = fmt.Errorf("get chain parameters error: %w", err)
		return
	}

	var (
		energyUnitPrice    int64 // getEnergyFee
		bandwidthUnitPrice int64 // getTransactionFee
	)
	// https://developers.tron.network/reference/wallet-getchainparameters
	slices.Values(params.GetChainParameter())(func(param *core.ChainParameters_ChainParameter) bool {
		if param != nil {
			switch param.Key {
			case "getEnergyFee":
				energyUnitPrice = param.Value
			case "getTransactionFee":
				bandwidthUnitPrice = param.Value
			}
			return energyUnitPrice == 0 || bandwidthUnitPrice == 0
		}
		return true
	})

	jsonStr := fmt.Sprintf(`[{"address":"%s"},{"uint256":"%s"}]`, to, amount)
	res, err := wc.cli.EstimateEnergy(wc.account, tokenAddress, "transfer(address,uint256)", jsonStr, 0, "", 0)
	if err != nil {
		return
	}
	energyFee := res.EnergyRequired * energyUnitPrice

	txExt, err := wc.cli.TRC20Send(wc.account, to, tokenAddress, amount, feeLimit)
	if err != nil {
		err = fmt.Errorf("create trc20 call tx error: %w", err)
		return
	}
	signature, err := crypto.Sign(txExt.Txid, wc.privateKey)
	if err != nil {
		err = fmt.Errorf("sign error: %w", err)
		return
	}
	txExt.Transaction.Signature = append(txExt.Transaction.Signature, signature)
	txBytes, err := proto.Marshal(txExt.Transaction)
	if err != nil {
		err = fmt.Errorf("marshal tx error: %w", err)
		return
	}
	// TRX burned to pay for Bandwidth = Total Bandwidth consumed by the transaction * Bandwidth unit price
	// https://developers.tron.network/docs/faq#5-how-to-calculate-the-bandwidth-and-energy-consumed-when-callingdeploying-a-contract
	bandwidthUsed := int64(len(txBytes) + 64)
	bandwidthFee := bandwidthUsed * bandwidthUnitPrice

	gas = Gas{
		Total:              energyFee + bandwidthFee,
		BandwidthUsed:      bandwidthUsed,
		BandwidthUnitPrice: bandwidthUnitPrice,
		EnergyUsed:         txExt.EnergyUsed,
		EnergyUnitPrice:    energyUnitPrice,
	}
	return
}

func (wc *WalletClient) EstimateGasTransferTRC20TokenV2(ctx context.Context, tokenAddress, to string, amount *big.Int, feeLimit int64) (gas Gas, err error) {
	params, err := wc.cli.Client.GetChainParameters(ctx, &api.EmptyMessage{})
	if err != nil {
		err = fmt.Errorf("get chain parameters error: %w", err)
		return
	}

	var (
		energyUnitPrice    int64 // getEnergyFee
		bandwidthUnitPrice int64 // getTransactionFee
	)
	// https://developers.tron.network/reference/wallet-getchainparameters
	slices.Values(params.GetChainParameter())(func(param *core.ChainParameters_ChainParameter) bool {
		if param != nil {
			switch param.Key {
			case "getEnergyFee":
				energyUnitPrice = param.Value
			case "getTransactionFee":
				bandwidthUnitPrice = param.Value
			}
			return energyUnitPrice == 0 || bandwidthUnitPrice == 0
		}
		return true
	})

	jsonStr := fmt.Sprintf(`[{"address":"%s"},{"uint256":"%s"}]`, to, amount)
	txExt, err := wc.cli.TriggerConstantContract(wc.account, tokenAddress, "transfer(address,uint256)", jsonStr)
	if err != nil {
		err = fmt.Errorf("TriggerConstantContract: %w", err)
		return
	}
	energyFee := txExt.EnergyUsed * energyUnitPrice

	tx := txExt.Transaction
	tx.RawData.FeeLimit = feeLimit
	tx.Ret = nil
	txRawDataBytes, err := proto.Marshal(tx.RawData)
	if err != nil {
		err = fmt.Errorf("marshal transaction raw data: %w", err)
		return
	}
	hash := sha256.Sum256(txRawDataBytes)
	signature, err := crypto.Sign(hash[:], wc.privateKey)
	if err != nil {
		err = fmt.Errorf("sign error: %w", err)
		return
	}
	tx.Signature = append(tx.Signature, signature)
	txBytes, err := proto.Marshal(tx)
	if err != nil {
		err = fmt.Errorf("marshal tx error: %w", err)
		return
	}
	// TRX burned to pay for Bandwidth = Total Bandwidth consumed by the transaction * Bandwidth unit price
	// https://developers.tron.network/docs/faq#5-how-to-calculate-the-bandwidth-and-energy-consumed-when-callingdeploying-a-contract
	bandwidthUsed := int64(len(txBytes) + 64)
	bandwidthFee := bandwidthUsed * bandwidthUnitPrice

	gas = Gas{
		Total:              energyFee + bandwidthFee,
		BandwidthUsed:      bandwidthUsed,
		BandwidthUnitPrice: bandwidthUnitPrice,
		EnergyUsed:         txExt.EnergyUsed,
		EnergyUnitPrice:    energyUnitPrice,
	}
	return
}

func (wc *WalletClient) TransferTRC20Token(ctx context.Context, tokenAddress, to string, amount *big.Int, feeLimit int64) (txHash string, err error) {
	txExt, err := wc.cli.TRC20Send(wc.account, to, tokenAddress, amount, feeLimit)
	if err != nil {
		err = fmt.Errorf("create trc20 call tx error: %w", err)
		return
	}
	signature, err := crypto.Sign(txExt.Txid, wc.privateKey)
	if err != nil {
		err = fmt.Errorf("sign error: %w", err)
		return
	}
	txExt.Transaction.Signature = append(txExt.Transaction.Signature, signature)
	ret, err := wc.cli.Broadcast(txExt.GetTransaction())
	if err != nil {
		err = fmt.Errorf("broadcast trx error: %v", err)
		return
	}
	if !ret.Result {
		err = fmt.Errorf("broadcast trx fail: %s", ret.String())
		return
	}
	txHash = hex.EncodeToString(txExt.GetTxid())
	return
}

func (wc *WalletClient) GetTRC20TokenBalance(ctx context.Context, tokenAddress string) (balance *big.Int, err error) {
	return wc.cli.TRC20ContractBalance(wc.account, tokenAddress)
}

func (wc *WalletClient) GetTRC20TokenBalanceByAddress(ctx context.Context, tokenAddress, address string) (balance *big.Int, err error) {
	return wc.cli.TRC20ContractBalance(address, tokenAddress)
}
