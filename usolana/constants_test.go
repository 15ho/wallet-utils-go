package usolana

import "os"

const (
	// Solana devnet USDC faucet: https://faucet.circle.com/
	USDCTokenAddress = "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU"
)

var (
	Acc1PrivateKeyBase58 = os.Getenv("ACC1PK58")
	Acc1AccountAddress   string

	Acc2PrivateKeyBase58 = os.Getenv("ACC2PK58")
	Acc2AccountAddress   string
)

func init() {
	if Acc1PrivateKeyBase58 != "" {
		addr, err := WalletAddressFromPrivateKey(Acc1PrivateKeyBase58)
		if err != nil {
			panic("get account1 address error: " + err.Error())
		}
		Acc1AccountAddress = addr
	}

	if Acc2PrivateKeyBase58 != "" {
		addr, err := WalletAddressFromPrivateKey(Acc2PrivateKeyBase58)
		if err != nil {
			panic("get account2 address error: " + err.Error())
		}
		Acc2AccountAddress = addr
	}
}
