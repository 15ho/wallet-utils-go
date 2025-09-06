package bip39

// BIP-0039: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
// This BIP describes the implementation of a mnemonic code or mnemonic sentence -- a group of easy to remember words -- for the generation of deterministic wallets.
// It consists of two parts: generating the mnemonic and converting it into a binary seed. This seed can be later used to generate deterministic wallets using BIP-0032 or similar methods.

// python implementation: https://github.com/trezor/python-mnemonic/tree/b57a5ad77a981e743f4167ab2f7927a55c1e82a8
