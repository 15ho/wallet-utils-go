package uethereum

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWalletAccount(t *testing.T) {
	privateKeyHex, address, err := CreateWalletAccount()
	assert.NoError(t, err)
	t.Logf("privateKeyHex: %s, address: %s", privateKeyHex, address)
}

func TestWalletClient(t *testing.T) {
	wc, err := NewWalletClient(EthTestnet, Acc1PrivateKeyHex)
	assert.NoError(t, err)
	ctx := t.Context()

	t.Run("estimate gas transfer eth", func(t *testing.T) {
		value, _ := big.NewInt(0).SetString("1000000000000", 10)
		gas, err := wc.EstimateGasTransferETH(ctx, Acc2AccountAddress, value)
		assert.NoError(t, err)
		t.Logf("gas: %d", gas)
	})

	t.Run("transfer eth", func(t *testing.T) {
		gasPrice, err := wc.SuggestGasPrice(ctx)
		assert.NoError(t, err)
		amount := new(big.Int).Div(WeiPerETH, big.NewInt(1000000)) // 0.000001 ETH
		txHash, err := wc.TransferETH(ctx, Acc2AccountAddress, amount, 100000, gasPrice)
		assert.NoError(t, err)
		t.Logf("txHash: %s", txHash)
	})

	t.Run("get eth balance", func(t *testing.T) {
		balance, err := wc.GetETHBalance(ctx)
		assert.NoError(t, err)
		t.Logf("acc1 eth balance: %s", balance)
	})

	t.Run("get eth balance by address", func(t *testing.T) {
		balance, err := wc.GetETHBalanceByAddress(ctx, Acc2AccountAddress)
		assert.NoError(t, err)
		t.Logf("acc2 eth balance: %s", balance)
	})

	t.Run("estimate transfer erc20 token", func(t *testing.T) {
		gas, err := wc.EstimateGasTransferERC20Token(ctx, USDCTokenAddress, Acc2AccountAddress, big.NewInt(1000000))
		assert.NoError(t, err)
		t.Logf("gas: %d", gas)
	})

	t.Run("transfer erc20 token", func(t *testing.T) {
		gasPrice, err := wc.SuggestGasPrice(ctx)
		assert.NoError(t, err)
		txHash, err := wc.TransferERC20Token(ctx, USDCTokenAddress, Acc2AccountAddress, big.NewInt(1), 100000, gasPrice)
		assert.NoError(t, err)
		t.Logf("txHash: %s", txHash)
	})

	t.Run("get erc20 token balance", func(t *testing.T) {
		balance, err := wc.GetERC20TokenBalance(ctx, USDCTokenAddress)
		assert.NoError(t, err)
		t.Logf("acc1 usdc balance: %s", balance)
	})

	t.Run("get erc20 token balance by address", func(t *testing.T) {
		balance, err := wc.GetERC20TokenBalanceByAddress(ctx, USDCTokenAddress, Acc2AccountAddress)
		assert.NoError(t, err)
		t.Logf("acc2 usdc balance: %s", balance)
	})
}
