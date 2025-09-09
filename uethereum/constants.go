package uethereum

import "math/big"

var (
	GweiPerETH = big.NewInt(1000000000)                               // 1 ETH = 1,000,000,000 Gwei
	WeiPerETH  = new(big.Int).Mul(GweiPerETH, big.NewInt(1000000000)) // 1 ETH = 1,000,000,000,000,000,000 Wei
)
