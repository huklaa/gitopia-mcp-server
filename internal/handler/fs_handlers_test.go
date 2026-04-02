package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gitopia/gitopia-mcp-server/internal/config"
	"github.com/gitopia/gitopia-mcp-server/internal/workspace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFile(t *testing.T) {
	tempDir := t.TempDir()

	wm, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tempDir})
	require.NoError(t, err)

	h := &ToolHandler{WorkspaceMgr: wm}

	// Write a file first
	testFilePath := "test-repo/test-file.txt"
	testContent := "hello, world!"
	fullPath := filepath.Join(tempDir, testFilePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
	require.NoError(t, os.WriteFile(fullPath, []byte(testContent), 0644))

	// Read the file
	result, _, err := h.ReadFile(context.Background(), &mcp.CallToolRequest{}, ReadFileParams{Path: testFilePath})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, testContent, result.Content[0].(*mcp.TextContent).Text)
}

func TestReadFile_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	wm, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tempDir})
	require.NoError(t, err)

	h := &ToolHandler{WorkspaceMgr: wm}

	result, _, err := h.ReadFile(context.Background(), &mcp.CallToolRequest{}, ReadFileParams{Path: "nonexistent.txt"})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "Failed to read")
}

func TestWriteFile(t *testing.T) {
	// Create a temporary directory for the workspace
	tempDir := t.TempDir()

	// Create a workspace manager for the temporary directory
	wm, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tempDir})
	require.NoError(t, err)

	// Create a tool handler with the workspace manager
	h := &ToolHandler{
		WorkspaceMgr: wm,
	}

	// Define the parameters for the WriteFile call
	testFilePath := "test-repo/test-file.txt"
	testContent := "hello, world!"
	params := WriteFileParams{
		Path:    testFilePath,
		Content: testContent,
	}

	// Call the WriteFile handler
	_, _, err = h.WriteFile(context.Background(), &mcp.CallToolRequest{}, params)
	require.NoError(t, err)

	// Verify that the file was written correctly
	fullPath := filepath.Join(tempDir, testFilePath)
	content, err := os.ReadFile(fullPath)
	require.NoError(t, err)

	assert.Equal(t, testContent, string(content))
}

func TestWriteFile_PathTraversal(t *testing.T) {
	tempDir := t.TempDir()

	wm, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tempDir})
	require.NoError(t, err)

	h := &ToolHandler{WorkspaceMgr: wm}

	result, _, err := h.WriteFile(context.Background(), &mcp.CallToolRequest{}, WriteFileParams{
		Path:    "../escape.txt",
		Content: "bad content",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "Failed to write")
}

func TestWriteFile_EmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tempDir})
	require.NoError(t, err)
	h := &ToolHandler{WorkspaceMgr: wm}

	_, _, err = h.WriteFile(context.Background(), &mcp.CallToolRequest{}, WriteFileParams{
		Path:    "empty.txt",
		Content: "",
	})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "empty.txt"))
	require.NoError(t, err)
	assert.Equal(t, "", string(content))
}

func TestReadFile_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tempDir})
	require.NoError(t, err)
	h := &ToolHandler{WorkspaceMgr: wm}

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "empty.txt"), []byte(""), 0644))

	result, _, err := h.ReadFile(context.Background(), &mcp.CallToolRequest{}, ReadFileParams{Path: "empty.txt"})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "", result.Content[0].(*mcp.TextContent).Text)
}

func TestReadFile_PathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tempDir})
	require.NoError(t, err)
	h := &ToolHandler{WorkspaceMgr: wm}

	result, _, err := h.ReadFile(context.Background(), &mcp.CallToolRequest{}, ReadFileParams{Path: "../../etc/passwd"})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
