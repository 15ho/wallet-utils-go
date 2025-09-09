package uethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func CreateWalletAccount() (privateKeyHex, address string, err error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return
	}
	pkBytes := crypto.FromECDSA(pk)
	privateKeyHex = hexutil.Encode(pkBytes)
	address = crypto.PubkeyToAddress(pk.PublicKey).Hex()
	return
}

type WalletClient struct {
	cli *ethclient.Client

	privateKey *ecdsa.PrivateKey
	account    common.Address
}

func NewWalletClient(endpoint, privateKeyHex string) (*WalletClient, error) {
	privateKey, err := crypto.ToECDSA(common.FromHex(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	cli, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}

	return &WalletClient{
		cli:        cli,
		privateKey: privateKey,
		account:    crypto.PubkeyToAddress(privateKey.PublicKey),
	}, nil
}

func (wc *WalletClient) EstimateGasTransferETH(ctx context.Context, to string, amount *big.Int) (gas uint64, err error) {
	toAddr := common.HexToAddress(to)
	gas, err = wc.cli.EstimateGas(ctx, ethereum.CallMsg{
		From:  wc.account,
		To:    &toAddr,
		Value: amount,
	})
	return
}

func (wc *WalletClient) SuggestGasPrice(ctx context.Context) (gasPrice *big.Int, err error) {
	return wc.cli.SuggestGasPrice(ctx)
}

func (wc *WalletClient) TransferETH(ctx context.Context, to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int) (txHash string, err error) {
	nonce, err := wc.cli.PendingNonceAt(ctx, wc.account)
	if err != nil {
		err = fmt.Errorf("get nonce: %v", err)
		return
	}
	chainID, err := wc.cli.ChainID(ctx)
	if err != nil {
		err = fmt.Errorf("get chain id: %v", err)
		return
	}
	tx := types.NewTransaction(nonce, common.HexToAddress(to), amount, gasLimit, gasPrice, nil)
	tx, err = types.SignTx(tx, types.NewEIP155Signer(chainID), wc.privateKey)
	if err != nil {
		err = fmt.Errorf("sign tx: %v", err)
		return
	}

	err = wc.cli.SendTransaction(ctx, tx)
	if err != nil {
		return
	}
	txHash = tx.Hash().Hex()
	return
}

func (wc *WalletClient) GetETHBalance(ctx context.Context) (balance *big.Int, err error) {
	return wc.cli.BalanceAt(ctx, wc.account, nil)
}

func (wc *WalletClient) GetETHBalanceByAddress(ctx context.Context, address string) (balance *big.Int, err error) {
	return wc.cli.BalanceAt(ctx, common.HexToAddress(address), nil)
}
