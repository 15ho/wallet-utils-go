package utron

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	tronaddr "github.com/fbsobreira/gotron-sdk/pkg/address"
)

func CreateWalletAccount() (privateKeyHex, address string, err error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return
	}
	pkBytes := crypto.FromECDSA(pk)
	privateKeyHex = hexutil.Encode(pkBytes)

	// base58(bytes(0x41) + common.Address.bytes())
	address = tronaddr.PubkeyToAddress(pk.PublicKey).String()
	return
}

func WalletAddressFromPrivateKey(privateKeyHex string) (address string, err error) {
	privateKeyBytes, err := hexutil.Decode(privateKeyHex)
	if err != nil {
		return
	}
	pk, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return
	}
	address = tronaddr.PubkeyToAddress(pk.PublicKey).String()
	return
}
