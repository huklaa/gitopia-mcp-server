package signing

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrGitopiaKeyNotConfigured(t *testing.T) {
	assert.True(t, errors.Is(ErrGitopiaKeyNotConfigured, ErrGitopiaKeyNotConfigured))
	assert.Contains(t, ErrGitopiaKeyNotConfigured.Error(), "not configured")
}

func TestInitWallet_MnemonicBackend(t *testing.T) {
	testMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	t.Setenv("GITOPIA_MNEMONIC", testMnemonic)

	// Use nil bank/feegrant clients since mnemonic path doesn't require balance check
	w, err := InitWallet(context.Background(), nil, nil, "mnemonic")
	require.NoError(t, err)
	assert.NotEmpty(t, w.Address())
}

func TestInitWallet_MnemonicBackend_NoEnv(t *testing.T) {
	t.Setenv("GITOPIA_MNEMONIC", "")

	_, err := InitWallet(context.Background(), nil, nil, "mnemonic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GITOPIA_MNEMONIC")
}

func TestInitWallet_UnknownBackend(t *testing.T) {
	_, err := InitWallet(context.Background(), nil, nil, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown wallet backend")
}
