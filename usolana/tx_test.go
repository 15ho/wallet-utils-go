package usolana

import (
	"encoding/json"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/assert"
)

func TestParseConfirmTx(t *testing.T) {
	tp := NewTxParser(rpc.MainNetBeta_RPC)
	tmw, err := tp.cli.GetTransaction(t.Context(),
		solana.MustSignatureFromBase58("4un2hKBgqTCTuVrUSPxSB6ivPfJ7AfV91pwx5CmUMy95EWENNRr3JPuhCs81NwgwaABWnBKAGdkhpuCf6pMZm68c"),
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase64,
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: rpc.NewTransactionVersion(0),
		})
	assert.NoError(t, err)
	ptx, err := tp.parseConfirmedTx(rpc.TransactionWithMeta{
		Slot:        tmw.Slot,
		BlockTime:   tmw.BlockTime,
		Transaction: rpc.DataBytesOrJSONFromBytes(tmw.Transaction.GetBinary()),
		Meta:        tmw.Meta,
		Version:     tmw.Version,
	})
	assert.NoError(t, err)
	t.Log("parsed tx:", ptx)
	ptxJson, err := json.Marshal(ptx)
	assert.NoError(t, err)
	t.Log("parsed tx json:", string(ptxJson))
}

// TODO: add more tests
