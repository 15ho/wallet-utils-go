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

func TestEstimateGasTransferETH(t *testing.T) {
	wc, err := NewWalletClient(EthTestnet, Acc1PrivateKeyHex)
	assert.NoError(t, err)
	ctx := t.Context()

	value, _ := big.NewInt(0).SetString("1000000000000", 10)
	gas, err := wc.EstimateGasTransferETH(ctx, Acc2AccountAddress, value)
	assert.NoError(t, err)
	t.Logf("gas: %d", gas)
}
