package bip39

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMnemonicGenerator(t *testing.T) {
	t.Run("valid language", func(t *testing.T) {
		for _, lang := range []string{
			"english",
			"chinese_simplified",
			"chinese_traditional",
			"czech",
			"french",
			"italian",
			"japanese",
			"korean",
			"portuguese",
			"russian",
			"spanish",
			"turkish",
		} {
			_, err := NewMnemonicGenerator(lang)
			assert.NoError(t, err)
		}
	})
	t.Run("invalid language", func(t *testing.T) {
		_, err := NewMnemonicGenerator("invalid")
		assert.Error(t, err)
	})
}

func TestMnemonicGenerator(t *testing.T) {
	lang := "english"
	mg, err := NewMnemonicGenerator(lang)
	assert.NoError(t, err)

	t.Run(lang, func(t *testing.T) {
		t.Run("generate mnemonic", func(t *testing.T) {
			words, err := mg.Generate()
			assert.NoError(t, err)
			assert.Len(t, strings.Split(words, " "), 12)
			t.Log(words)

			for _, strength := range []int{128, 160, 192, 224, 256} {
				words, err := mg.Generate(strength)
				assert.NoError(t, err)
				assert.Len(t, strings.Split(words, " "), strength/32*3)
				t.Log(words)
			}
		})
	})

	lang = "japanese"
	mg, err = NewMnemonicGenerator(lang)
	assert.NoError(t, err)
	t.Run(lang, func(t *testing.T) {
		t.Run("generate mnemonic", func(t *testing.T) {
			words, err := mg.Generate()
			assert.NoError(t, err)
			assert.Len(t, strings.Split(words, "\u3000"), 12)
			t.Log(words)
		})
	})
}

func TestCreateSeedFromMnemonic(t *testing.T) {
	mg, err := NewMnemonicGenerator("english")
	assert.NoError(t, err)

	t.Run("create", func(t *testing.T) {
		words, err := mg.Generate()
		assert.NoError(t, err)
		t.Log(words)
		seed := CreateSeedFromMnemonic(words)
		assert.Len(t, seed, 64)
		t.Log(seed, hex.EncodeToString(seed))
		seed = CreateSeedFromMnemonic(words, "12347890")
		assert.Len(t, seed, 64)
		t.Log(seed, hex.EncodeToString(seed))
	})

	t.Run("validate", func(t *testing.T) {
		words := "prefer judge blouse motor naive october legal labor exact sustain stuff direct"
		expectSeed1, _ := hex.DecodeString("e2e576197f191c9ac9a13af747c9ee3104701a37a6f018438ec827147cfd674f6279adf0a209b8e7e3f6c7bd20e0be1ba6d20055aced88b5030296e1bdecb044")
		expectSeed2, _ := hex.DecodeString("2ffd59fae330185352dd5079121a7957513eef7cad19c873e391341e8ec1b07ff4d2d07115f697460d65317b54f6e4de53f624fbb9c9009dcddab02551c27333")
		seed2Passphrase := "151515888"

		seed1 := CreateSeedFromMnemonic(words)
		seed2 := CreateSeedFromMnemonic(words, seed2Passphrase)
		assert.EqualValues(t, expectSeed1, seed1)
		assert.EqualValues(t, expectSeed2, seed2)
	})

}
