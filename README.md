# Wallet Utils Library for Go

[![Tests on Linux, MacOS and Windows](https://github.com/15ho/wallet-utils-go/workflows/CI/badge.svg)](https://github.com/15ho/wallet-utils-go/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/15ho/wallet-utils-go/branch/main/graph/badge.svg)](https://codecov.io/gh/15ho/wallet-utils-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/15ho/wallet-utils-go)](https://goreportcard.com/report/github.com/15ho/wallet-utils-go)

> NOTE: This project is under development.

## Features

- Generate wallet private key and wallet address
- Construct a transfer transaction
- Estimate transfer transaction fees
- Parse transaction information in a block
- L2 Token Bridging

## Support Chains

> 🛠️ The features are primarily built using the Chain SDK.

||Chain |SDK |License |Description |
|:------------:|:---------:|:--------:|:--------:|:---------:|
|<img src="https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/solana/info/logo.png" width="32" /> |Solana |[solana-go](https://github.com/gagliardetto/solana-go) |Apache-2.0 |Go SDK library and RPC client for the Solana Blockchain
|<img src="https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/info/logo.png" width="32" /> |Ethereum |[go-ethereum/ethclient](https://github.com/ethereum/go-ethereum) |LGPL-3.0, GPL-3.0 |Go implementation of the Ethereum protocol
|<img src="https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/tron/info/logo.png" width="32" /> |Tron |[gotron-sdk](https://github.com/fbsobreira/gotron-sdk) |LGPL-3.0 |Tron SDK for golang / CLI tool with keystore manager
|<img src="https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/arbitrum/info/logo.png" width="32" /> |Arbitrum |[arbitrum-sdk](https://github.com/OffchainLabs/arbitrum-sdk) |Apache-2.0 license |A TypeScript library for client-side interactions with Arbitrum