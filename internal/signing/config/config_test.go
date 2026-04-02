package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitConfigGet_InvalidKey(t *testing.T) {
	// A nonexistent git config key should return an error
	_, err := GitConfigGet("nonexistent_key_that_does_not_exist")
	assert.Error(t, err)
}

func TestDefaultValues(t *testing.T) {
	assert.Equal(t, "gitopia", ChainId)
	assert.Equal(t, "ulore", Denom)
	assert.Equal(t, "0.001ulore", GasPrices)
	assert.Equal(t, "https://server.gitopia.com", GitServerHost)
	assert.NotEmpty(t, FeeGranterAddr)
}

func TestLoadGitConfig_NoGitConfig(t *testing.T) {
	// Save original values
	origChainId := ChainId
	origDenom := Denom
	origGasPrices := GasPrices
	origGitServerHost := GitServerHost
	origFeeGranterAddr := FeeGranterAddr
	defer func() {
		ChainId = origChainId
		Denom = origDenom
		GasPrices = origGasPrices
		GitServerHost = origGitServerHost
		FeeGranterAddr = origFeeGranterAddr
	}()

	// LoadGitConfig should not panic even if git config doesn't have our section.
	// It may fail silently on individual keys, but returns nil.
	err := LoadGitConfig()
	assert.NoError(t, err)
}
