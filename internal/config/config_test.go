package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Workspace)
	assert.NotNil(t, cfg.Git)
	assert.NotNil(t, cfg.Logging)
	assert.NotNil(t, cfg.RateLimit)
	assert.Equal(t, "chainwrite", cfg.TrustLevel)
	assert.Equal(t, 10, cfg.RateLimit.ChainTxPerMinute)
	assert.Equal(t, 100, cfg.RateLimit.ChainTxPerHour)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Empty(t, cfg.Workspace.BasePath)
}

func TestLoadConfig_NoFile(t *testing.T) {
	cfg, err := LoadConfig("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "chainwrite", cfg.TrustLevel)
}

func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.json")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	data := []byte(`{
		"trust_level": "readonly",
		"logging": {"level": "debug"},
		"rate_limit": {"chain_tx_per_minute": 5, "chain_tx_per_hour": 50}
	}`)
	require.NoError(t, os.WriteFile(cfgPath, data, 0644))

	cfg, err := LoadConfig(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "readonly", cfg.TrustLevel)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, 5, cfg.RateLimit.ChainTxPerMinute)
	assert.Equal(t, 50, cfg.RateLimit.ChainTxPerHour)
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	require.NoError(t, os.WriteFile(cfgPath, []byte("not json"), 0644))

	_, err := LoadConfig(cfgPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse config")
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("TRUST_LEVEL", "readonly")
	t.Setenv("MCP_LOG_LEVEL", "debug")

	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "readonly", cfg.TrustLevel)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestLoadConfigFromEnv_Transport(t *testing.T) {
	t.Setenv("TRANSPORT", "http")
	t.Setenv("PORT", "9090")

	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "http", cfg.Transport)
	assert.Equal(t, "9090", cfg.HTTPPort)
}

func TestLoadConfigFromEnv_TransportDefaults(t *testing.T) {
	t.Setenv("TRANSPORT", "")
	t.Setenv("PORT", "")

	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "", cfg.Transport) // empty means stdio (default)
	assert.Equal(t, "", cfg.HTTPPort)  // empty means 8080 (default applied at runtime)
}

func TestLoadConfigFromEnv_ApprovalMode(t *testing.T) {
	t.Setenv("APPROVAL_MODE", "true")
	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.True(t, cfg.ApprovalMode)
}

func TestLoadConfigFromEnv_ApprovalMode_Numeric(t *testing.T) {
	t.Setenv("APPROVAL_MODE", "1")
	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.True(t, cfg.ApprovalMode)
}

func TestLoadConfigFromEnv_ApprovalMode_False(t *testing.T) {
	t.Setenv("APPROVAL_MODE", "false")
	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.False(t, cfg.ApprovalMode)
}

func TestLoadConfigFromEnv_BooleanOverridesCanDisableFileValues(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{
		"dry_run": true,
		"approval_mode": true
	}`), 0644))

	t.Setenv("MCP_CONFIG_FILE", cfgPath)
	t.Setenv("DRY_RUN", "false")
	t.Setenv("APPROVAL_MODE", "false")

	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.False(t, cfg.DryRun)
	assert.False(t, cfg.ApprovalMode)
}

func TestLoadConfigFromEnv_ApprovalTTL(t *testing.T) {
	t.Setenv("APPROVAL_TTL", "10m")
	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "10m", cfg.ApprovalTTL)
}

func TestLoadConfigFromEnv_Toolsets(t *testing.T) {
	t.Setenv("TOOLSETS", "core")
	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "core", cfg.Toolsets)
}

func TestLoadConfigFromEnv_ToolsetsDefault(t *testing.T) {
	cfg, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.Empty(t, cfg.Toolsets) // empty means "all"
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "subdir", "config.json")

	cfg := DefaultConfig()
	require.NoError(t, cfg.SaveConfig(cfgPath))

	// Verify file was created
	_, err := os.Stat(cfgPath)
	assert.NoError(t, err)

	// Verify it can be loaded back
	loaded, err := LoadConfig(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, cfg.TrustLevel, loaded.TrustLevel)
}
