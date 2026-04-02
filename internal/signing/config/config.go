package config

// Config package for Gitopia signing utilities.
// Provides default settings that may be overridden via git config or environment variables.

import (
    "fmt"
    "os/exec"
    "strings"
)

// Default configuration keys and values.
const (
    AppName = "git-remote-gitopia"

    // git config section and option keys used by the Gitopia CLI / tooling.
    GitopiaConfigSection             = "gitopia"
    GitopiaConfigChainIdOption       = "chainId"
    GitopiaConfigGRPCHostOption      = "grpcHost"
    GitopiaConfigTmAddrOption        = "tmAddr"
    GitopiaConfigGitServerHostOption = "gitServerHost"
    GitopiaConfigKeyOption           = "key"
    GitopiaConfigBackendOption       = "backend"
    GitopiaConfigGasPricesOption     = "gasPrices"
    GitopiaConfigFeeGranterOption    = "feeGranter"
    GitopiaConfigDenomOption         = "denom"
)

// Mutable values that may be overridden at runtime by LoadGitConfig.
var (
	ChainId        = "gitopia"
	GitServerHost  = "https://server.gitopia.com"
	GasPrices      = "0.001ulore"
	FeeGranterAddr = "gitopia13ashgc6j5xle4m47kqyn5psavq0u3klmscfxql"
	Denom          = "ulore"
)

// GitConfigGet reads a value from global git config: `git config --get gitopia.<key>`.
func GitConfigGet(key string) (string, error) {
    cmd := exec.Command("git", "config", "--get", fmt.Sprintf("gitopia.%s", key))
    stdout, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(stdout)), nil
}

// LoadGitConfig populates the package variables from the user's git config if they exist.
func LoadGitConfig() error {
    if res, err := GitConfigGet(GitopiaConfigChainIdOption); err == nil && res != "" {
        ChainId = res
    }
    if res, err := GitConfigGet(GitopiaConfigGitServerHostOption); err == nil && res != "" {
        GitServerHost = res
    }
    if res, err := GitConfigGet(GitopiaConfigGasPricesOption); err == nil && res != "" {
        GasPrices = res
    }
    if res, err := GitConfigGet(GitopiaConfigFeeGranterOption); err == nil && res != "" {
        FeeGranterAddr = res
    }
    if res, err := GitConfigGet(GitopiaConfigDenomOption); err == nil && res != "" {
        Denom = res
    }
    return nil
}
