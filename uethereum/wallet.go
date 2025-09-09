package uethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func (wc *WalletClient) EstimateGasTransferETH(ctx context.Context, to string, value *big.Int) (gas uint64, err error) {
	toAddr := common.HexToAddress(to)
	gas, err = wc.cli.EstimateGas(ctx, ethereum.CallMsg{
		From:  wc.account,
		To:    &toAddr,
		Value: value,
	})
	return
}
