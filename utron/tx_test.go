package utron

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxParser(t *testing.T) {
	tp, cleanup, err := NewTxParser(TronMainnet)
	assert.NoError(t, err)
	defer cleanup()

	txCases := []struct {
		Name   string
		TxHash string
	}{
		{
			"transfer trx tx",
			"6cbbe4ef8747c9a5688eb6be62c4aba0bb2a3c5ec3833be78772560ca3830374",
		},
		{
			"transfer trc20 tx",
			"bc95f5fe79d26d06ea61067af9078520e13a5b7e88b5f790694277fb616f5e0d",
		},
		{
			"delegate resource",
			"00ab875ce610f3473470f09d06aea364ecbccc960049c43f753645852ac1aab2",
		},
		{
			"reclaim resource",
			"12edfc95b06d655c01e0cff35eeb2c3fcfcb358ea7037ffd9d786cc47ab691f9",
		},
	}
	slices.Values(txCases)(func(tc struct {
		Name   string
		TxHash string
	}) bool {
		t.Run(tc.Name, func(t *testing.T) {
			txHash := tc.TxHash
			txExt, err := tp.cli.GetTransactionByID(txHash)
			assert.NoError(t, err)
			t.Logf("txExt: %+v", txExt)

			txInfo, err := tp.cli.GetTransactionInfoByID(txHash)
			assert.NoError(t, err)
			t.Logf("txInfo: %+v", txInfo)

			ptx, err := tp.parseTx(txExt, txInfo)
			assert.NoError(t, err)
			t.Logf("parsedTx: %+v", ptx)
		})
		return true
	})

}
