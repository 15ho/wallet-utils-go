package utron

import "os"

const (
	// tron testnet USDT faucet: https://nileex.io/join/getJoinPage
	USDTTokenAddress = "TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf"
)

var (
	Acc1PrivateKeyHex  = os.Getenv("ACC1PKHEX")
	Acc1AccountAddress string

	Acc2PrivateKeyHex  = os.Getenv("ACC2PKHEX")
	Acc2AccountAddress string

	TronTestnet = os.Getenv("TRONTESTNET")
	TronMainnet = os.Getenv("TRONMAINNET")
)

func init() {
	if Acc1PrivateKeyHex != "" {
		addr, err := WalletAddressFromPrivateKey(Acc1PrivateKeyHex)
		if err != nil {
			panic("get account1 address error: " + err.Error())
		}
		Acc1AccountAddress = addr
	}

	if Acc2PrivateKeyHex != "" {
		addr, err := WalletAddressFromPrivateKey(Acc2PrivateKeyHex)
		if err != nil {
			panic("get account2 address error: " + err.Error())
		}
		Acc2AccountAddress = addr
	}
}
