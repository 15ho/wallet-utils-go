package usolana

import (
	"errors"
	"math/big"
	"testing"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/assert"
)

func TestCreateWalletAccount(t *testing.T) {
	pk58, addr, err := CreateWalletAccount()
	assert.NoError(t, err)
	assert.NotEmpty(t, pk58)
	assert.NotEmpty(t, addr)
	t.Logf("private key base58: %s, account address: %s", pk58, addr)
}

func TestWalletClient(t *testing.T) {
	if Acc1PrivateKeyBase58 == "" {
		t.Skip("ACC1PK58 env var is not set")
	}
	if Acc2PrivateKeyBase58 == "" {
		t.Skip("ACC2PK58 env var is not set")
	}

	wc, err := NewDevnetWalletClient(Acc1PrivateKeyBase58)
	assert.NoError(t, err)
	ctx := t.Context()

	t.Run("get sol balance by address", func(t *testing.T) {
		balance, err := wc.GetSOLBalanceByAddress(ctx, Acc2AccountAddress)
		assert.NoError(t, err)
		t.Logf("account2 balance: %d", balance)
	})

	t.Run("get sol balance", func(t *testing.T) {
		balance, err := wc.GetSOLBalance(ctx)
		assert.NoError(t, err)
		t.Logf("account1 balance: %d", balance)
	})

	t.Run("transfer sol", func(t *testing.T) {
		amount := LamportsPerSOL / 1000000
		_, err := wc.GetSOLBalanceByAddress(ctx, Acc2AccountAddress)
		if err != nil && !errors.Is(err, rpc.ErrNotFound) {
			assert.NoError(t, err)
		}
		if err != nil {
			// Accounts require a rent deposit in lamports (SOL) that's proportional to the amount of data stored,
			// and you can fully recover it when you close the account.
			// https://solana.com/docs/core/accounts#rent
			// transfer sol: data size is 0
			rentAmount, err := wc.cli.GetMinimumBalanceForRentExemption(ctx, 0, rpc.CommitmentFinalized)
			assert.NoError(t, err)
			amount = rentAmount
			t.Logf("rent amount: %d", rentAmount)
		}

		balance, err := wc.GetSOLBalance(ctx)
		assert.NoError(t, err)
		if balance < amount+5000 {
			t.Skip("balance is not enough")
		}
		sign, err := wc.TransferSOL(ctx, Acc2AccountAddress, amount) // 0.000001 SOL
		assert.NoError(t, err)
		t.Logf("signature: %s", sign)
	})

	t.Run("get spl token balance by address", func(t *testing.T) {
		balance, decimals, err := wc.GetSPLTokenBalanceByAddress(ctx, USDCTokenAddress, Acc2AccountAddress)
		assert.NoError(t, err)
		t.Logf("account2 spl token balance: %s, decimals: %d", balance, decimals)
	})

	t.Run("get spl token balance", func(t *testing.T) {
		balance, decimals, err := wc.GetSPLTokenBalance(ctx, USDCTokenAddress)
		assert.NoError(t, err)
		t.Logf("account1 spl token balance: %s, decimals: %d", balance, decimals)
	})

	t.Run("transfer spl token", func(t *testing.T) {
		amount := uint64(1)
		splTokenBalance, _, err := wc.GetSPLTokenBalance(ctx, USDCTokenAddress)
		assert.NoError(t, err)
		balance, _ := new(big.Int).SetString(splTokenBalance, 10)
		assert.NotNil(t, balance)
		if balance.Cmp(big.NewInt(int64(amount))) < 0 {
			t.Skip("balance is not enough")
		}

		sign, err := wc.TransferSPLToken(ctx, USDCTokenAddress, Acc2AccountAddress, amount)
		assert.NoError(t, err)
		t.Logf("signature: %s", sign)
	})
}
