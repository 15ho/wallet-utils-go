package uethereum

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestTxParser(t *testing.T) {
	wc, err := NewTxParser(EthMainnet)
	assert.NoError(t, err)
	ctx := t.Context()

	txHash := common.HexToHash("0xde06bb51d04c738853d5f1b555857503d01ef16fec4b2e5cc5aa1a0a8ef65a95")
	tx, isPending, err := wc.cli.TransactionByHash(ctx, txHash)
	assert.NoError(t, err)
	assert.False(t, isPending)

	r, err := wc.cli.TransactionReceipt(ctx, txHash)
	assert.NoError(t, err)

	h, err := wc.cli.HeaderByNumber(ctx, r.BlockNumber)
	assert.NoError(t, err)

	ptx, err := wc.parseTx(h, tx, r)
	assert.NoError(t, err)
	t.Logf("parsed tx: %+v", ptx)
	for _, log := range ptx.Logs {
		t.Logf("parsed log: %+v", log)
	}
}

// TODO: add more tests
