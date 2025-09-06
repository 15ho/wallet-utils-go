package uethereum

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func CreateWalletAccount() (privateKeyHex, address string, err error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return
	}
	pkBytes := crypto.FromECDSA(pk)
	privateKeyHex = hexutil.Encode(pkBytes)
	address = crypto.PubkeyToAddress(pk.PublicKey).Hex()
	return
}
