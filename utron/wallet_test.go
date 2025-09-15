package utron

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWalletAccount(t *testing.T) {
	privateKeyHex, address, err := CreateWalletAccount()
	assert.NoError(t, err)
	t.Logf("privateKeyHex: %s, address: %s", privateKeyHex, address)

	address2, err := WalletAddressFromPrivateKey(privateKeyHex)
	assert.NoError(t, err)
	t.Logf("address2: %s", address2)
}

func TestWalletClient(t *testing.T) {
	wc, cleanup, err := NewWalletClient(TronTestnet, Acc1PrivateKeyHex)
	assert.NoError(t, err)
	defer cleanup()
	ctx := t.Context()

	t.Run("estimate gas transfer trx", func(t *testing.T) {
		gas, err := wc.EstimateGasTransferTRX(ctx, Acc2AccountAddress, SunPerTRX/1000)
		assert.NoError(t, err)
		t.Logf("gas: %d", gas)
	})

	t.Run("estimate gas transfer trx to inactivated account", func(t *testing.T) {
		_, address, err := CreateWalletAccount()
		assert.NoError(t, err)
		gas, err := wc.EstimateGasTransferTRX(ctx, address, SunPerTRX/1000)
		assert.NoError(t, err)
		t.Logf("gas: %d", gas)
	})

	t.Run("transfer trx", func(t *testing.T) {
		txHash, err := wc.TransferTRX(ctx, Acc2AccountAddress, SunPerTRX/1000)
		assert.NoError(t, err)
		t.Logf("txHash: %s", txHash)
	})

	t.Run("get trx balance", func(t *testing.T) {
		balance, err := wc.GetTRXBalance(ctx)
		assert.NoError(t, err)
		t.Logf("acc1 trx balance: %d", balance)
	})

	t.Run("get trx balance by address", func(t *testing.T) {
		balance, err := wc.GetTRXBalanceByAddress(ctx, Acc2AccountAddress)
		assert.NoError(t, err)
		t.Logf("acc2 trx balance: %d", balance)
	})

	t.Run("transfer trc20 token", func(t *testing.T) {
		txHash, err := wc.TransferTRC20Token(ctx, USDTTokenAddress, Acc2AccountAddress, big.NewInt(1), 100*SunPerTRX)
		assert.NoError(t, err)
		t.Logf("txHash: %s", txHash)
	})

	t.Run("get trc20 token balance", func(t *testing.T) {
		balance, err := wc.GetERC20TokenBalance(ctx, USDTTokenAddress)
		assert.NoError(t, err)
		t.Logf("acc1 trc20 token balance: %d", balance)
	})

	t.Run("get trc20 token balance by address", func(t *testing.T) {
		balance, err := wc.GetERC20TokenBalanceByAddress(ctx, USDTTokenAddress, Acc2AccountAddress)
		assert.NoError(t, err)
		t.Logf("acc2 trc20 token balance: %d", balance)
	})
}
