package usolana

import (
	"testing"

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

	wc, err := NewTestnetWalletClient(Acc1PrivateKeyBase58)
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
		balance, err := wc.GetSOLBalance(ctx)
		assert.NoError(t, err)
		if balance < amount*10 {
			t.Skip("balance is not enough")
		}
		sign, err := wc.TransferSOL(ctx, Acc2AccountAddress, amount) // 0.000001 SOL
		assert.NoError(t, err)
		t.Logf("signature: %s", sign)
	})
}
