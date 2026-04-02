package workspace

import (
	"testing"

	"github.com/gitopia/gitopia-mcp-server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager_NilConfig(t *testing.T) {
	_, err := NewManager(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestNewManager_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&config.WorkspaceConfig{
		BasePath: tmpDir,
	})
	require.NoError(t, err)
	assert.NotNil(t, mgr)
	assert.Equal(t, tmpDir, mgr.GetBasePath())
}

func TestResolvePath_Simple(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&config.WorkspaceConfig{
		BasePath: tmpDir,
	})
	require.NoError(t, err)

	absPath, err := mgr.ResolvePath("myrepo")
	require.NoError(t, err)
	assert.Contains(t, absPath, "myrepo")
}

func TestResolvePath_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&config.WorkspaceConfig{
		BasePath: tmpDir,
	})
	require.NoError(t, err)

	_, err = mgr.ResolvePath("../etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

func TestResolvePath_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&config.WorkspaceConfig{
		BasePath: tmpDir,
	})
	require.NoError(t, err)

	_, err = mgr.ResolvePath("/etc/passwd")
	assert.Error(t, err)
}

func TestResolvePath_NestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&config.WorkspaceConfig{
		BasePath: tmpDir,
	})
	require.NoError(t, err)

	absPath, err := mgr.ResolvePath("repo/src/main.go")
	require.NoError(t, err)
	assert.Contains(t, absPath, "repo/src/main.go")
}
