package uarbitrum

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalletClient(t *testing.T) {
	wc, err := NewWalletClient(EthTestnet, ArbTestnet, Acc1PrivateKeyHex)
	assert.NoError(t, err)
	ctx := t.Context()

	t.Run("approve l1 token", func(t *testing.T) {
		gasPrice, err := wc.l1cli.SuggestGasPrice(ctx)
		assert.NoError(t, err)

		txHash, bridgeGatewayAddr, err := wc.ApproveL1Token(ctx, USDCTokenAddress, L1GatewayRouterAddress, big.NewInt(100), 100000, gasPrice)
		assert.NoError(t, err)
		assert.Equal(t, L1ERC20GatewayAddress, bridgeGatewayAddr)
		t.Logf("tx hash: %s", txHash)
	})

	t.Run("deposit l1 token", func(t *testing.T) {
		txHash, err := wc.DepositL1Token(ctx, USDCTokenAddress, L1GatewayRouterAddress,
			big.NewInt(100), big.NewInt(334172200303680),
			248965, big.NewInt(1500000000), big.NewInt(1500000000),
			big.NewInt(540287), big.NewInt(600000000), big.NewInt(189800),
		)
		assert.NoError(t, err)
		t.Logf("tx hash: %s", txHash)
	})
}
