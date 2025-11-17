package uarbitrum

import (
	"os"

	"github.com/15ho/wallet-utils-go/uethereum"
)

const (
	// Ethereum testnet USDC faucet: https://faucet.circle.com/
	USDCTokenAddress = "0x1c7d4b196cb0c7b01d743fbc6116a902379c7238"

	// https://docs.arbitrum.io/build-decentralized-apps/reference/contract-addresses#token-bridge-smart-contracts
	L1GatewayRouterAddress = "0xcE18836b233C83325Cc8848CA4487e94C6288264"
	L1ERC20GatewayAddress = "0x902b3E5f8F19571859F4AB1003B960a5dF693aFF"
)

var (
	Acc1PrivateKeyHex  = os.Getenv("ACC1PKHEX")
	Acc1AccountAddress string

	EthTestnet = os.Getenv("ETHTESTNET")
	ArbTestnet = os.Getenv("ARBTESTNET")
)

func init() {
	if Acc1PrivateKeyHex != "" {
		addr, err := uethereum.WalletAddressFromPrivateKey(Acc1PrivateKeyHex)
		if err != nil {
			panic("get account1 address error: " + err.Error())
		}
		Acc1AccountAddress = addr
	}

}