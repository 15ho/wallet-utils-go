package uethereum

import "os"

var (
	Acc1PrivateKeyHex  = os.Getenv("ACC1PKHEX")
	Acc1AccountAddress = os.Getenv("ACC1ADDRHEX")

	Acc2PrivateKeyHex  = os.Getenv("ACC2PKHEX")
	Acc2AccountAddress = os.Getenv("ACC2ADDRHEX")

	EthTestnet = os.Getenv("ETHTESTNET")
)
