package utron

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	tronaddr "github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
		grpc.WithPerRPCCredentials(auth{}))
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

func (wc *WalletClient) GetERC20TokenBalance(ctx context.Context, tokenAddress string) (balance *big.Int, err error) {
	return wc.cli.TRC20ContractBalance(wc.account, tokenAddress)
}

func (wc *WalletClient) GetERC20TokenBalanceByAddress(ctx context.Context, tokenAddress, address string) (balance *big.Int, err error) {
	return wc.cli.TRC20ContractBalance(address, tokenAddress)
}
