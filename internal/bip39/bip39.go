package bip39

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/15ho/wallet-utils-go/internal/bip39/wordlist"
)

// BIP-0039: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
// This BIP describes the implementation of a mnemonic code or mnemonic sentence -- a group of easy to remember words -- for the generation of deterministic wallets.
// It consists of two parts: generating the mnemonic and converting it into a binary seed. This seed can be later used to generate deterministic wallets using BIP-0032 or similar methods.

// python implementation: https://github.com/trezor/python-mnemonic/tree/b57a5ad77a981e743f4167ab2f7927a55c1e82a8

type MnemonicGenerator struct {
	wordList  []string
	delimiter string
}

// NewMnemonicGenerator returns a new MnemonicGenerator for the given language.
func NewMnemonicGenerator(language string) (*MnemonicGenerator, error) {
	var wordListStr string
	mnemonicSep := " "
	switch strings.ToLower(language) {
	case "english":
		wordListStr = wordlist.English
	case "chinese_simplified":
		wordListStr = wordlist.ChineseSimplified
	case "chinese_traditional":
		wordListStr = wordlist.ChineseTraditional
	case "spanish":
		wordListStr = wordlist.Spanish
	case "french":
		wordListStr = wordlist.French
	case "italian":
		wordListStr = wordlist.Italian
	case "japanese":
		wordListStr = wordlist.Japanese
		mnemonicSep = "\u3000"
	case "korean":
		wordListStr = wordlist.Korean
	case "portuguese":
		wordListStr = wordlist.Portuguese
	case "russian":
		wordListStr = wordlist.Russian
	case "turkish":
		wordListStr = wordlist.Turkish
	case "czech":
		wordListStr = wordlist.Czech
	default:
		return nil, fmt.Errorf("language %s not supported", language)
	}

	wordList := strings.Split(strings.TrimSpace(wordListStr), "\n")
	if len(wordList) != 2048 {
		return nil, fmt.Errorf("invalid word list length: %d", len(wordList))
	}
	return &MnemonicGenerator{
		wordList:  wordList,
		delimiter: mnemonicSep,
	}, nil
}

// Generate Create a new mnemonic using a random generated number as entropy.
// As defined in BIP39, the entropy must be a multiple of 32 bits, and its size must be between 128 and 256 bits.
// Therefore the possible values for `strength` are 128, 160, 192, 224 and 256.
// If not provided, the default entropy length will be set to 128 bits.
// The return is a list of words that encodes the generated entropy.
func (mg *MnemonicGenerator) Generate(strengthOption ...int) (string, error) {
	strength := 128
	if len(strengthOption) > 0 {
		strength = strengthOption[0]
	}
	if !slices.Contains([]int{128, 160, 192, 224, 256}, strength) {
		return "", fmt.Errorf("invalid strength: %d", strength)
	}
	data := make([]byte, strength/8)
	_, _ = rand.Read(data)
	return mg.toMnemonic(data)
}

func (mg *MnemonicGenerator) toMnemonic(data []byte) (string, error) {
	if !slices.Contains([]int{16, 20, 24, 28, 32}, len(data)) {
		return "", fmt.Errorf("invalid data length: %d", len(data))
	}
	h256 := sha256.Sum256(data)
	h := hex.EncodeToString(h256[:])

	b1 := fmt.Sprintf("%0*b", len(data)*8, new(big.Int).SetBytes(data))

	hInt, _ := new(big.Int).SetString(h, 16)
	b2 := fmt.Sprintf("%0*b", 256, hInt)[:len(data)*8/32]

	b := b1 + b2
	words := make([]string, len(b)/11)
	for i := 0; i < len(words); i++ {
		start := i * 11
		end := start + 11
		bInt, _ := new(big.Int).SetString(b[start:end], 2)
		words[i] = mg.wordList[bInt.Int64()]
	}
	return strings.Join(words, mg.delimiter), nil
}
