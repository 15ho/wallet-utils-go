package usolana

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/gagliardetto/solana-go"
	tokenacc "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
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

func NewDevnetWalletClient(privateKeyBase58 string) (*WalletClient, error) {
	return NewWalletClient(rpc.DevNet_RPC, privateKeyBase58)
}

func (wc *WalletClient) buildTxTransferSOL(ctx context.Context, toAddress string, amount uint64, priorityFeeOption ...TxPriorityFee) (tx *solana.Transaction, err error) {
	to, err := solana.PublicKeyFromBase58(toAddress)
	if err != nil {
		return
	}

	res, err := wc.cli.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		err = fmt.Errorf("get latest block hash: %w", err)
		return
	}

	inss := make([]solana.Instruction, 0, 3)

	if len(priorityFeeOption) > 0 {
		priorityFee := priorityFeeOption[0]
		inss = append(inss,
			computebudget.NewSetComputeUnitLimitInstruction(priorityFee.ComputeUnitLimit).Build(),
			computebudget.NewSetComputeUnitPriceInstruction(priorityFee.ComputeUnitPrice).Build(),
		)
	}

	tx, err = solana.NewTransaction(
		append(inss, system.NewTransferInstruction(
			amount,
			wc.account,
			to,
		).Build()),
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
	}
	return
}

func (wc *WalletClient) SimulateTxTransferSOL(ctx context.Context, toAddress string, amount uint64, priorityFeeOption ...TxPriorityFee) (unitsConsumed uint64, err error) {
	tx, err := wc.buildTxTransferSOL(ctx, toAddress, amount, priorityFeeOption...)
	if err != nil {
		return
	}
	simRes, err := wc.cli.SimulateTransaction(ctx, tx)
	if err != nil {
		err = fmt.Errorf("simulate tx: %w", err)
		return
	}
	unitsConsumed = *simRes.Value.UnitsConsumed
	return
}

func (wc *WalletClient) TransferSOL(ctx context.Context, toAddress string, amount uint64, priorityFeeOption ...TxPriorityFee) (signature string, err error) {
	tx, err := wc.buildTxTransferSOL(ctx, toAddress, amount, priorityFeeOption...)
	if err != nil {
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

func (wc *WalletClient) buildTxTransferSPLToken(ctx context.Context, tokenAddress, toAddress string, amount uint64, priorityFeeOption ...TxPriorityFee) (tx *solana.Transaction, err error) {
	mint, err := solana.PublicKeyFromBase58(tokenAddress)
	if err != nil {
		err = fmt.Errorf("parse mint: %w", err)
		return
	}
	to, err := solana.PublicKeyFromBase58(toAddress)
	if err != nil {
		err = fmt.Errorf("parse to address: %w", err)
		return
	}
	fromTokenAcc, _, err := solana.FindAssociatedTokenAddress(wc.account, mint)
	if err != nil {
		err = fmt.Errorf("find associated from token account: %w", err)
		return
	}
	toTokenAcc, _, err := solana.FindAssociatedTokenAddress(to, mint)
	if err != nil {
		err = fmt.Errorf("find associated to token account: %w", err)
		return
	}

	accRes, err := wc.cli.GetAccountInfo(ctx, toTokenAcc)
	if err != nil && err != rpc.ErrNotFound {
		err = fmt.Errorf("get to token account info: %w", err)
		return
	}

	inss := make([]solana.Instruction, 0, 4)

	if len(priorityFeeOption) > 0 {
		priorityFee := priorityFeeOption[0]
		inss = append(inss,
			computebudget.NewSetComputeUnitLimitInstruction(priorityFee.ComputeUnitLimit).Build(),
			computebudget.NewSetComputeUnitPriceInstruction(priorityFee.ComputeUnitPrice).Build(),
		)
	}

	if accRes == nil {
		// create spl token account
		// https://solana.com/docs/tokens#token-account
		// https://solana.com/developers/cookbook/tokens/create-token-account
		inss = append(inss,
			tokenacc.NewCreateInstruction(
				wc.account,
				to,
				mint,
			).Build(),
		)
	}

	res, err := wc.cli.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		err = fmt.Errorf("get latest block hash: %w", err)
		return
	}
	tx, err = solana.NewTransaction(
		append(inss, token.NewTransferInstruction(amount,
			fromTokenAcc,
			toTokenAcc,
			wc.account,
			nil,
		).Build()),
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
	}
	return
}

func (wc *WalletClient) SimulateTxTransferSPLToken(ctx context.Context, tokenAddress, toAddress string, amount uint64, priorityFeeOption ...TxPriorityFee) (unitsConsumed uint64, err error) {
	tx, err := wc.buildTxTransferSPLToken(ctx, tokenAddress, toAddress, amount, priorityFeeOption...)
	if err != nil {
		return
	}

	simRes, err := wc.cli.SimulateTransaction(ctx, tx)
	if err != nil {
		err = fmt.Errorf("simulate tx: %w", err)
		return
	}
	unitsConsumed = *simRes.Value.UnitsConsumed
	return
}

func (wc *WalletClient) TransferSPLToken(ctx context.Context, tokenAddress, toAddress string, amount uint64, priorityFeeOption ...TxPriorityFee) (signature string, err error) {
	tx, err := wc.buildTxTransferSPLToken(ctx, tokenAddress, toAddress, amount, priorityFeeOption...)
	if err != nil {
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

func (wc *WalletClient) GetSPLTokenBalance(ctx context.Context, tokenAddress string) (balance string, decimals uint8, err error) {
	mint, err := solana.PublicKeyFromBase58(tokenAddress)
	if err != nil {
		err = fmt.Errorf("parse mint: %w", err)
		return
	}
	return wc.getSPLTokenBalance(ctx, mint, wc.account)
}

func (wc *WalletClient) getSPLTokenBalance(ctx context.Context, mint, walletAccount solana.PublicKey) (balance string, decimals uint8, err error) {
	tokenAcc, _, err := solana.FindAssociatedTokenAddress(walletAccount, mint)
	if err != nil {
		err = fmt.Errorf("find associated token account: %w", err)
		return
	}
	res, err := wc.cli.GetTokenAccountBalance(ctx, tokenAcc, rpc.CommitmentFinalized)
	if err != nil {
		err = fmt.Errorf("get token account balance: %w", err)
		return
	}
	balance = res.Value.Amount
	decimals = res.Value.Decimals
	return
}

func (wc *WalletClient) GetSPLTokenBalanceByAddress(ctx context.Context, tokenAddress, walletAddress string) (balance string, decimals uint8, err error) {
	mint, err := solana.PublicKeyFromBase58(tokenAddress)
	if err != nil {
		err = fmt.Errorf("parse mint: %w", err)
		return
	}
	account, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		err = fmt.Errorf("parse wallet address: %w", err)
		return
	}
	return wc.getSPLTokenBalance(ctx, mint, account)
}
