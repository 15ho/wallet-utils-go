package usolana

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
)

func CreateWalletAccount() (privateKeyBase58, address string, err error) {
	_, pk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	privateKeyBase58 = base58.Encode(pk)
	address = base58.Encode(pk.Public().(ed25519.PublicKey))
	return
}

type WalletClient struct {
	cli *rpc.Client

	privateKey solana.PrivateKey
	account    solana.PublicKey
}

func NewWalletClient(endpoint, privateKeyBase58 string) (*WalletClient, error) {
	privateKey, err := solana.PrivateKeyFromBase58(privateKeyBase58)
	if err != nil {
		return nil, err
	}
	return &WalletClient{
		cli:        rpc.New(endpoint),
		privateKey: privateKey,
		account:    privateKey.PublicKey(), // NOTE: Check if the provided `b` is on the ed25519 curve.
	}, nil
}

func NewTestnetWalletClient(privateKeyBase58 string) (*WalletClient, error) {
	return NewWalletClient(rpc.TestNet_RPC, privateKeyBase58)
}

func (wc *WalletClient) TransferSOL(ctx context.Context, toAddress string, amount uint64) (signature string, err error) {
	to, err := solana.PublicKeyFromBase58(toAddress)
	if err != nil {
		return
	}

	res, err := wc.cli.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		err = fmt.Errorf("get latest block hash: %w", err)
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				amount,
				wc.account,
				to,
			).Build(),
		},
		res.Value.Blockhash,
		solana.TransactionPayer(wc.account),
	)
	if err != nil {
		err = fmt.Errorf("new tx: %w", err)
		return
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if wc.account.Equals(key) {
				return &wc.privateKey
			}
			return nil
		},
	)
	if err != nil {
		err = fmt.Errorf("tx sign: %w", err)
		return
	}

	signObj, err := wc.cli.SendTransaction(ctx, tx)
	if err != nil {
		err = fmt.Errorf("send tx: %w", err)
		return
	}
	signature = signObj.String()
	return
}

func (wc *WalletClient) GetSOLBalance(ctx context.Context) (balance uint64, err error) {
	return wc.getSOLBalance(ctx, wc.account)
}

func (wc *WalletClient) getSOLBalance(ctx context.Context, account solana.PublicKey) (balance uint64, err error) {
	res, err := wc.cli.GetAccountInfo(ctx, account)
	if err != nil {
		return
	}
	return res.Value.Lamports, nil
}

func (wc *WalletClient) GetSOLBalanceByAddress(ctx context.Context, address string) (balance uint64, err error) {
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return
	}
	return wc.getSOLBalance(ctx, pubKey)
}
