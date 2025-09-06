package uethereum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWalletAccount(t *testing.T) {
	privateKeyHex, address, err := CreateWalletAccount()
	assert.NoError(t, err)
	t.Logf("privateKeyHex: %s, address: %s", privateKeyHex, address)
}
