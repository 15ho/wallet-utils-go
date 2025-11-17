package uarbitrum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type WalletClient struct {
	l1cli *ethclient.Client
	l2cli *ethclient.Client

	privateKey *ecdsa.PrivateKey
	account    common.Address
}

func NewWalletClient(ethEndpoint, arbEndpoint, privateKeyHex string) (*WalletClient, error) {
	privateKey, err := crypto.ToECDSA(common.FromHex(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	l1cli, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, fmt.Errorf("L1 dial: %v", err)
	}

	l2cli, err := ethclient.Dial(arbEndpoint)
	if err != nil {
		return nil, fmt.Errorf("L2 dial: %v", err)
	}

	return &WalletClient{
		l1cli:      l1cli,
		l2cli:      l2cli,
		privateKey: privateKey,
		account:    crypto.PubkeyToAddress(privateKey.PublicKey),
	}, nil
}

func (wc *WalletClient) ApproveL1Token(ctx context.Context,
	tokenAddress, l1GatewayRouterAddress string,
	amount *big.Int,
	gasLimit uint64, gasPrice *big.Int) (txHash, l1GatewayAddress string, err error) {

	tokenAddr := common.HexToAddress(tokenAddress)
	l1GatewayRouterAddr := common.HexToAddress(l1GatewayRouterAddress)

	data, err := l1GatewayRouterABI.Pack("getGateway", tokenAddr)
	if err != nil {
		err = fmt.Errorf("l1GatewayRouterABI pack err: %w", err)
		return
	}
	l1GatewayAddrBytes, err := wc.l1cli.CallContract(ctx, ethereum.CallMsg{
		To:   &l1GatewayRouterAddr,
		Data: data,
	}, nil)
	if err != nil {
		err = fmt.Errorf("call contract: getGateway: %w", err)
		return
	}

	l1GatewayAddr := common.BytesToAddress(l1GatewayAddrBytes)
	l1GatewayAddress = l1GatewayAddr.String()

	nonce, err := wc.l1cli.PendingNonceAt(ctx, wc.account)
	if err != nil {
		err = fmt.Errorf("get nonce: %w", err)
		return
	}
	chainID, err := wc.l1cli.ChainID(ctx)
	if err != nil {
		err = fmt.Errorf("get chain id: %w", err)
		return
	}
	approveData, err := erc20ABI.Pack("approve", l1GatewayAddr, amount)
	if err != nil {
		err = fmt.Errorf("erc20ABI pack err: %w", err)
		return
	}
	tx := types.NewTransaction(nonce, tokenAddr, nil, gasLimit, gasPrice, approveData)
	tx, err = types.SignTx(tx, types.NewEIP155Signer(chainID), wc.privateKey)
	if err != nil {
		err = fmt.Errorf("sign tx: %w", err)
		return
	}
	err = wc.l1cli.SendTransaction(ctx, tx)
	if err != nil {
		err = fmt.Errorf("send tx: %w", err)
		return
	}
	txHash = tx.Hash().Hex()
	return
}

// TODO: calculate gas and fee
// https://docs.arbitrum.io/how-arbitrum-works/deep-dives/gas-and-fees
// - l1tol2Fee
// - maxGas
// - gasPriceBid
// - maxSubmissionCost

func (wc *WalletClient) DepositL1Token(ctx context.Context,
	tokenAddress, l1GatewayRouterAddress string,
	amount, l1tol2Fee *big.Int,
	gasLimit uint64, maxFeePerGas, maxPriorityFeePerGas *big.Int,
	maxGas, gasPriceBid, maxSubmissionCost *big.Int) (txHash string, err error) {
	nonce, err := wc.l1cli.PendingNonceAt(ctx, wc.account)
	if err != nil {
		err = fmt.Errorf("get nonce: %w", err)
		return
	}
	chainID, err := wc.l1cli.ChainID(ctx)
	if err != nil {
		err = fmt.Errorf("get chain id: %w", err)
		return
	}
	l1GatewayRouterAddr := common.HexToAddress(l1GatewayRouterAddress)
	outboundTransferData, err := outboundTransferDataArgs.Pack(maxSubmissionCost, []byte{})
	if err != nil {
		err = fmt.Errorf("outboundTransferDataArgs pack: %w", err)
		return
	}
	data, err := l1GatewayRouterABI.Pack("outboundTransfer", common.HexToAddress(tokenAddress), wc.account, amount, maxGas, gasPriceBid, outboundTransferData)
	if err != nil {
		err = fmt.Errorf("l1GatewayRouterABI pack err: %w", err)
		return
	}
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		Gas:       gasLimit,
		GasFeeCap: maxFeePerGas,
		GasTipCap: maxPriorityFeePerGas,
		To:        &l1GatewayRouterAddr,
		Value:     l1tol2Fee,
		Data:      data,
	})
	tx, err = types.SignTx(tx, types.NewLondonSigner(chainID), wc.privateKey)
	if err != nil {
		err = fmt.Errorf("sign tx: %w", err)
		return
	}
	err = wc.l1cli.SendTransaction(ctx, tx)
	if err != nil {
		err = fmt.Errorf("send tx: %w", err)
		return
	}
	txHash = tx.Hash().Hex()
	return
}
