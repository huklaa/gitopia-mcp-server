package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gitopia/gitopia-mcp-server/internal/config"
	"github.com/gitopia/gitopia-mcp-server/internal/constants"
)

// Manager handles workspace operations and path resolution
type Manager struct {
	baseDir string
}

// NewManager creates a new workspace manager
func NewManager(cfg *config.WorkspaceConfig) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("workspace config cannot be nil")
	}

	baseDir, err := resolveBasePath(cfg.BasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	// Ensure the workspace directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory %s: %w", baseDir, err)
	}

	return &Manager{baseDir: baseDir}, nil
}

// resolveBasePath resolves the base path for the workspace.
// Priority: explicit config path > PWD (if it's a git repo) > default (~/.mcp/gitopia/workspace).
func resolveBasePath(configPath string) (string, error) {
	var basePath string

	if configPath != "" {
		basePath = expandPath(configPath)
	} else if pwd, err := os.Getwd(); err == nil {
		// If launched from a directory with a .git folder, use it as workspace.
		// This makes native binary usage from an editor "just work" without
		// needing to set MCP_WORKSPACE_PATH.
		if _, statErr := os.Stat(filepath.Join(pwd, ".git")); statErr == nil {
			basePath = pwd
		}
	}

	if basePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		basePath = filepath.Join(homeDir, constants.GetMCPGitopiaWorkspacePath())
	}

	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	path = os.ExpandEnv(path)

	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	return path
}

// ResolvePath resolves a relative path to an absolute workspace path.
// Automatically creates parent directories as needed.
func (m *Manager) ResolvePath(relativePath string) (string, error) {
	cleanedPath := filepath.Clean(relativePath)

	if filepath.IsAbs(cleanedPath) || strings.HasPrefix(cleanedPath, "..") {
		return "", fmt.Errorf("invalid path: %s", relativePath)
	}

	absPath := filepath.Join(m.baseDir, cleanedPath)

	rel, err := filepath.Rel(m.baseDir, absPath)
	if err != nil {
		return "", fmt.Errorf("could not create relative path: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path is outside of workspace: %s", relativePath)
	}

	m.ensureParentDir(absPath)

	return absPath, nil
}

// ensureParentDir ensures the parent directory of a path exists
func (m *Manager) ensureParentDir(fullPath string) {
	parentDir := filepath.Dir(fullPath)
	if parentDir != "" && parentDir != "." && parentDir != "/" {
		_ = os.MkdirAll(parentDir, 0755)
	}
}

// GetBasePath returns the base workspace path
func (m *Manager) GetBasePath() string {
	return m.baseDir
}
