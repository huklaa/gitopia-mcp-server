package signing

import (
	"strings"
	"testing"

	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Standard BIP-39 test vector mnemonic
const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func TestFromMnemonic_Valid(t *testing.T) {
	w, err := FromMnemonic(testMnemonic)
	require.NoError(t, err)
	assert.NotEmpty(t, w.Address())
	assert.Equal(t, ENV_VAR, w.Type())
}

func TestFromMnemonic_TooShort(t *testing.T) {
	_, err := FromMnemonic("too short mnemonic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestFromMnemonic_Empty(t *testing.T) {
	_, err := FromMnemonic("")
	assert.Error(t, err)
}

func TestFromMnemonic_Whitespace(t *testing.T) {
	// Leading/trailing whitespace should be trimmed
	w, err := FromMnemonic("  " + testMnemonic + "  ")
	require.NoError(t, err)
	assert.NotEmpty(t, w.Address())
}

func TestFromMnemonic_AddressDeterministic(t *testing.T) {
	w1, err := FromMnemonic(testMnemonic)
	require.NoError(t, err)

	w2, err := FromMnemonic(testMnemonic)
	require.NoError(t, err)

	assert.Equal(t, w1.Address(), w2.Address())
}

func TestGenerateMnemonic(t *testing.T) {
	m, err := GenerateMnemonic()
	require.NoError(t, err)

	words := strings.Fields(m)
	assert.Len(t, words, 24, "expected 24-word mnemonic")
	assert.True(t, bip39.IsMnemonicValid(m), "generated mnemonic should be valid BIP-39")
}

func TestGenerateMnemonic_Unique(t *testing.T) {
	m1, err := GenerateMnemonic()
	require.NoError(t, err)

	m2, err := GenerateMnemonic()
	require.NoError(t, err)

	assert.NotEqual(t, m1, m2, "two generated mnemonics should differ")
}
