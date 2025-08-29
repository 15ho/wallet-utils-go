package usolana

import "os"

var (
	Acc1PrivateKeyBase58 = os.Getenv("ACC1PK58")
	Acc1AccountAddress   = os.Getenv("ACC1ADDR")

	Acc2PrivateKeyBase58 = os.Getenv("ACC2PK58")
	Acc2AccountAddress   = os.Getenv("ACC2ADDR")
)

const (
	// Solana devnet USDC faucet: https://faucet.circle.com/
	USDCTokenAddress = "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU"
)
