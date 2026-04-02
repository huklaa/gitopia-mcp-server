package signing

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/gitopia/gitopia-mcp-server/internal/constants"
)

const (
	keyringService = "gitopia-mcp-server"
	keyringKey     = "gitopia-wallet"
)

// ErrGitopiaKeyNotConfigured is returned when no mnemonic is found in the OS keyring.
var ErrGitopiaKeyNotConfigured = errors.New("gitopia key not configured in OS keyring")

// FromKeyring opens the OS keyring, retrieves the stored mnemonic, and
// constructs a Wallet with feegrant support.
func FromKeyring(ctx context.Context, bankClient banktypes.QueryClient, feegrantClient feegrant.QueryClient) (Wallet, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyringService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open OS keyring: %w", err)
	}

	item, err := ring.Get(keyringKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return nil, ErrGitopiaKeyNotConfigured
		}
		return nil, fmt.Errorf("failed to read from OS keyring: %w", err)
	}

	mnemonic := strings.TrimSpace(string(item.Data))
	if mnemonic == "" {
		return nil, ErrGitopiaKeyNotConfigured
	}

	w, err := FromMnemonicWithFeegrant(ctx, mnemonic, bankClient, feegrantClient)
	if err != nil {
		// Chain query failed (e.g. new account, unreachable node) — still
		// return a usable wallet without feegrant info.
		return FromMnemonic(mnemonic)
	}
	return w, nil
}

const walletFileName = "wallet.key"

// walletFilePath returns the absolute path to the wallet key file:
// ~/.mcp/gitopia/config/wallet.key
func walletFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, constants.GetMCPGitopiaConfigPath(), walletFileName), nil
}

// StoreInFile persists a mnemonic to the wallet key file with secure permissions.
func StoreInFile(mnemonic string) error {
	path, err := walletFilePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create wallet directory %s: %w", dir, err)
	}
	if err := os.WriteFile(path, []byte(mnemonic), 0600); err != nil {
		return fmt.Errorf("failed to write wallet file %s: %w", path, err)
	}
	return nil
}

// FromFile reads the mnemonic from the wallet key file and constructs a Wallet.
func FromFile(ctx context.Context, bankClient banktypes.QueryClient, feegrantClient feegrant.QueryClient) (Wallet, error) {
	path, err := walletFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}
	mnemonic := strings.TrimSpace(string(data))
	if mnemonic == "" {
		return nil, fmt.Errorf("wallet file %s is empty", path)
	}
	w, err := FromMnemonicWithFeegrant(ctx, mnemonic, bankClient, feegrantClient)
	if err != nil {
		return FromMnemonic(mnemonic)
	}
	return w, nil
}

// StoreInKeyring stores a mnemonic in the OS keyring under the gitopia service.
func StoreInKeyring(mnemonic string) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyringService,
	})
	if err != nil {
		return fmt.Errorf("failed to open OS keyring: %w", err)
	}

	return ring.Set(keyring.Item{
		Key:         keyringKey,
		Data:        []byte(mnemonic),
		Label:       "Gitopia MCP Server Wallet",
		Description: "BIP-39 mnemonic for Gitopia MCP server",
	})
}

// InitWallet creates a Wallet using the specified backend.
// Supported backends: "keyring", "mnemonic", "file", "auto" (or empty).
// The "auto" backend tries: env var -> wallet key file -> legacy wallet file -> auto-generate.
// OS keyring is only used when explicitly set via GITOPIA_WALLET_BACKEND=keyring.
func InitWallet(ctx context.Context, bankClient banktypes.QueryClient, feegrantClient feegrant.QueryClient, backend string) (Wallet, error) {
	switch strings.ToLower(strings.TrimSpace(backend)) {
	case "keyring":
		return FromKeyring(ctx, bankClient, feegrantClient)
	case "mnemonic":
		mnemonic := os.Getenv("GITOPIA_MNEMONIC")
		if mnemonic == "" {
			return nil, fmt.Errorf("GITOPIA_MNEMONIC env var is required for mnemonic backend")
		}
		return FromMnemonicWithFeegrant(ctx, mnemonic, bankClient, feegrantClient)
	case "file":
		return FromFile(ctx, bankClient, feegrantClient)
	case "auto", "":
		// 1. Try mnemonic env var (fast path, no I/O)
		if mnemonic := os.Getenv("GITOPIA_MNEMONIC"); mnemonic != "" {
			return FromMnemonicWithFeegrant(ctx, mnemonic, bankClient, feegrantClient)
		}
		// 2. Try wallet key file (no OS popup)
		w, err := FromFile(ctx, bankClient, feegrantClient)
		if err == nil {
			return w, nil
		}
		// 3. Try legacy wallet file (GITOPIA_WALLET)
		w, err = InitGitopiaWallet(ctx, bankClient, feegrantClient)
		if err == nil {
			return w, nil
		}
		// 5. Auto-generate a new wallet as last resort
		mnemonic, err := GenerateMnemonic()
		if err != nil {
			return nil, fmt.Errorf("failed to auto-generate wallet: %w", err)
		}
		// Persist to file BEFORE use (crash safety, no popup)
		if storeErr := StoreInFile(mnemonic); storeErr != nil {
			return nil, fmt.Errorf("failed to store auto-generated wallet: %w", storeErr)
		}
		w, err = FromMnemonic(mnemonic)
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet from generated mnemonic: %w", err)
		}
		// Mark as auto-generated so callers can detect first-use
		if gw, ok := w.(*GitopiaWallet); ok {
			gw.autoGenerated = true
			return gw, nil
		}
		return w, nil
	default:
		return nil, fmt.Errorf("unknown wallet backend %q: use keyring, mnemonic, file, or auto", backend)
	}
}
