package bip39

import (
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
