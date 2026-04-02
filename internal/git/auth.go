package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// AuthManager handles git authentication for Gitopia
type AuthManager struct {
	WalletPath string
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{}
}

// SetupAuth sets up git authentication using a wallet
func (am *AuthManager) SetupAuth(walletData []byte, walletAddress string) error {
	// Create temporary wallet file
	tempDir := os.TempDir()
	walletFileName := fmt.Sprintf("gitopia-wallet-%s.json", walletAddress)
	walletPath := filepath.Join(tempDir, walletFileName)

	// Write wallet data to temporary file with secure permissions
	if err := os.WriteFile(walletPath, walletData, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	// Set environment variable for git-remote-gitopia
	if err := os.Setenv("GITOPIA_WALLET", walletPath); err != nil {
		_ = os.Remove(walletPath)
		return fmt.Errorf("set GITOPIA_WALLET: %w", err)
	}

	am.WalletPath = walletPath
	return nil
}

// CleanupAuth removes the temporary wallet file and unsets environment variables
func (am *AuthManager) CleanupAuth() error {
	var errors []error

	// Remove temporary wallet file
	if am.WalletPath != "" {
		if err := os.Remove(am.WalletPath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("failed to remove wallet file: %w", err))
		}
		am.WalletPath = ""
	}

	// Unset environment variable
	if err := os.Unsetenv("GITOPIA_WALLET"); err != nil {
		errors = append(errors, fmt.Errorf("failed to unset GITOPIA_WALLET: %w", err))
	}

	// Return first error if any
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// IsAuthSetup checks if git authentication is properly configured
func (am *AuthManager) IsAuthSetup() bool {
	walletPath := os.Getenv("GITOPIA_WALLET")
	if walletPath == "" {
		return false
	}

	// Check if wallet file exists
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetWalletPath returns the current wallet path
func (am *AuthManager) GetWalletPath() string {
	return os.Getenv("GITOPIA_WALLET_PATH")
}

// SetupGitConfig sets the global git user name and email.
func SetupGitConfig(ctx context.Context, name, email string) error {
	if err := runGitGlobalConfig(ctx, "user.name", name); err != nil {
		return err
	}
	if err := runGitGlobalConfig(ctx, "user.email", email); err != nil {
		return err
	}
	return nil
}

// runGitGlobalConfig executes a git config --global command.
func runGitGlobalConfig(ctx context.Context, key, value string) error {
	c := NewClient()
	_, err := c.runGitCommand(ctx, "", "config", "--global", key, value)
	return err
}

// ValidateGitopiaURL checks if a URL is a valid Gitopia URL
func ValidateGitopiaURL(url string) bool {
	return len(url) > 10 && (url[:10] == "gitopia://" ||
		(len(url) > 8 && url[:8] == "https://" && contains(url, "gitopia.com")))
}

// NormalizeGitopiaURL converts various Gitopia URL formats to the canonical gitopia:// format
func NormalizeGitopiaURL(url string) string {
	// If already gitopia://, return as-is
	if len(url) > 10 && url[:10] == "gitopia://" {
		return url
	}

	// Convert https://gitopia.com/owner/repo to gitopia://owner/repo
	if len(url) > 8 && url[:8] == "https://" && contains(url, "gitopia.com") {
		// Extract owner/repo from https://gitopia.com/owner/repo
		parts := splitURL(url)
		if len(parts) >= 2 {
			return fmt.Sprintf("gitopia://%s/%s", parts[0], parts[1])
		}
	}

	// If it looks like owner/repo format, convert to gitopia://
	if !contains(url, "://") && contains(url, "/") {
		return fmt.Sprintf("gitopia://%s", url)
	}

	return url
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func splitURL(url string) []string {
	// Simple URL parsing for gitopia.com URLs
	// Expected format: https://gitopia.com/owner/repo
	if !contains(url, "gitopia.com/") {
		return nil
	}

	// Find the part after gitopia.com/
	start := findSubstring(url, "gitopia.com/")
	if start == -1 {
		return nil
	}

	path := url[start+len("gitopia.com/"):]

	// Split by '/' and return first two parts
	var parts []string
	current := ""
	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
