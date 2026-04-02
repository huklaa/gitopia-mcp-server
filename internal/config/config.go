package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gitopia/gitopia-mcp-server/internal/constants"
)

// RateLimitConfig holds rate limiting configuration for chain transactions.
type RateLimitConfig struct {
	ChainTxPerMinute int `json:"chain_tx_per_minute"`
	ChainTxPerHour   int `json:"chain_tx_per_hour"`
}

// Config represents the complete MCP server configuration
type Config struct {
	Workspace  *WorkspaceConfig `json:"workspace,omitempty"`
	Git        *GitConfig       `json:"git,omitempty"`
	Logging    *LoggingConfig   `json:"logging,omitempty"`
	TrustLevel string           `json:"trust_level,omitempty"`
	RateLimit  *RateLimitConfig `json:"rate_limit,omitempty"`
	DryRun         bool             `json:"dry_run,omitempty"`
	GRPCEndpoints  []string         `json:"grpc_endpoints,omitempty"`
	Transport      string           `json:"transport,omitempty"`
	HTTPPort       string           `json:"http_port,omitempty"`
	ApprovalMode   bool             `json:"approval_mode,omitempty"`
	ApprovalTTL    string           `json:"approval_ttl,omitempty"`
	Toolsets       string           `json:"toolsets,omitempty"`
}

// WorkspaceConfig holds workspace-specific configuration.
// This is defined here to avoid circular dependencies.
type WorkspaceConfig struct {
	BasePath string `json:"base_path"`
}

// GitConfig holds git-specific configuration
type GitConfig struct {
	ConfigIsolation   bool   `json:"config_isolation"`
	CredentialHelper  string `json:"credential_helper"`
	DefaultUserName   string `json:"default_user_name"`
	DefaultUserEmail  string `json:"default_user_email"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level    string `json:"level"`
	FilePath string `json:"file_path"`
	MaxSize  int    `json:"max_size_mb"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		TrustLevel: "chainwrite",
		RateLimit: &RateLimitConfig{
			ChainTxPerMinute: 10,
			ChainTxPerHour:   100,
		},
		Workspace: &WorkspaceConfig{
			BasePath: "", // Let the workspace manager resolve the default
		},
		Git: &GitConfig{
			ConfigIsolation:   true,
			CredentialHelper:  "mcp-gitopia",
			DefaultUserName:   "Gitopia MCP Server",
			DefaultUserEmail:  "mcp@gitopia.com",
		},
		Logging: &LoggingConfig{
			Level:   "info",
			MaxSize: 10,
		},
	}
}

// LoadConfig loads configuration from a file, with fallback to defaults
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if configPath == "" {
		return config, nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use defaults
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON and merge with defaults
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// LoadConfigFromEnv loads configuration from the default or specified config file
// and overlays environment variable overrides.
func LoadConfigFromEnv() (*Config, error) {
    // Determine config path (MCP_CONFIG_FILE takes precedence; otherwise default path)
    cfgPath := GetConfigPath()

    // Load file-based config (falls back to defaults if file does not exist)
    config, err := LoadConfig(cfgPath)
    if err != nil {
        return nil, err
    }

    // Overlay with environment variables
    if workspacePath := os.Getenv("MCP_WORKSPACE_PATH"); workspacePath != "" {
        config.Workspace.BasePath = workspacePath
    }

    if gitName := os.Getenv("GIT_USER_NAME"); gitName != "" {
        config.Git.DefaultUserName = gitName
    }

    if gitEmail := os.Getenv("GIT_USER_EMAIL"); gitEmail != "" {
        config.Git.DefaultUserEmail = gitEmail
    }

    if logLevel := os.Getenv("MCP_LOG_LEVEL"); logLevel != "" {
        config.Logging.Level = logLevel
    }

    if trustLevel := os.Getenv("TRUST_LEVEL"); trustLevel != "" {
        config.TrustLevel = trustLevel
    }

    if dryRun := os.Getenv("DRY_RUN"); dryRun == "true" || dryRun == "1" {
        config.DryRun = true
    }

    if endpoints := os.Getenv("GITOPIA_GRPC_ENDPOINTS"); endpoints != "" {
        config.GRPCEndpoints = strings.Split(endpoints, ",")
        for i := range config.GRPCEndpoints {
            config.GRPCEndpoints[i] = strings.TrimSpace(config.GRPCEndpoints[i])
        }
    }

    if transport := os.Getenv("TRANSPORT"); transport != "" {
        config.Transport = transport
    }

    if port := os.Getenv("PORT"); port != "" {
        config.HTTPPort = port
    }

    if v := os.Getenv("APPROVAL_MODE"); v == "true" || v == "1" {
        config.ApprovalMode = true
    }

    if v := os.Getenv("APPROVAL_TTL"); v != "" {
        config.ApprovalTTL = v
    }

    if v := os.Getenv("TOOLSETS"); v != "" {
        config.Toolsets = v
    }

    return config, nil
}

// SaveConfig saves the configuration to a file
func (c *Config) SaveConfig(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	if configPath := os.Getenv("MCP_CONFIG_FILE"); configPath != "" {
		return configPath
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "mcp-config.json" // fallback to local file
	}

	return filepath.Join(homeDir, constants.GetMCPGitopiaConfigPath(), constants.ConfigFileName)
}
