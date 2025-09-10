package uethereum

import "os"

var (
	Acc1PrivateKeyHex  = os.Getenv("ACC1PKHEX")
	Acc1AccountAddress = os.Getenv("ACC1ADDRHEX")

	Acc2PrivateKeyHex  = os.Getenv("ACC2PKHEX")
	Acc2AccountAddress = os.Getenv("ACC2ADDRHEX")

	EthTestnet = os.Getenv("ETHTESTNET")
)

const (
	// Ethereum testnet USDC faucet: https://faucet.circle.com/
	USDCTokenAddress = "0x1c7d4b196cb0c7b01d743fbc6116a902379c7238"
)
