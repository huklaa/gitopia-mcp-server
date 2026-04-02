package handler

import (
	"context"
	"testing"

	"github.com/gitopia/gitopia-mcp-server/internal/config"
	"github.com/gitopia/gitopia-mcp-server/internal/git"
	"github.com/gitopia/gitopia-mcp-server/internal/workspace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGitTestHandler(t *testing.T) *ToolHandler {
	t.Helper()
	tempDir := t.TempDir()
	wm, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tempDir})
	require.NoError(t, err)
	return &ToolHandler{
		WorkspaceMgr: wm,
		GitClient:    git.NewClient(),
	}
}

// ---- GitClone Tests ----

func TestGitClone_NoWallet(t *testing.T) {
	h := newGitTestHandler(t)
	h.GClient = nil // no gitopia client -> wallet fails
	result, _, err := h.GitClone(context.Background(), &mcp.CallToolRequest{}, GitCloneParams{
		RepoURL: "gitopia://test/repo", LocalPath: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Equal(t, "error", m["status"])
	assert.Contains(t, m["message"], "Authentication failed")
}

// ---- GitStatus Tests ----

func TestGitStatus_NotARepo(t *testing.T) {
	h := newGitTestHandler(t)
	result, _, err := h.GitStatus(context.Background(), &mcp.CallToolRequest{}, GitStatusParams{
		RepoPath: "nonexistent-repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Equal(t, "NOT_REPOSITORY", m["code"])
}

// ---- GitAdd Tests ----

func TestGitAdd_NotARepo(t *testing.T) {
	h := newGitTestHandler(t)
	result, _, err := h.GitAdd(context.Background(), &mcp.CallToolRequest{}, GitAddParams{
		RepoPath: "nonexistent-repo", Files: []string{"file.txt"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Equal(t, "NOT_REPOSITORY", m["code"])
}

// ---- GitCommit Tests ----

func TestGitCommit_NotARepo(t *testing.T) {
	h := newGitTestHandler(t)
	result, _, err := h.GitCommit(context.Background(), &mcp.CallToolRequest{}, GitCommitParams{
		RepoPath: "nonexistent-repo", Message: "test commit",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Equal(t, "NOT_REPOSITORY", m["code"])
}

// ---- GitPush Tests ----

func TestGitPush_NoWallet(t *testing.T) {
	h := newGitTestHandler(t)
	h.GClient = nil
	result, _, err := h.GitPush(context.Background(), &mcp.CallToolRequest{}, GitPushParams{
		RepoPath: "some-repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Contains(t, m["message"], "Authentication failed")
}

// ---- CreateFeatureBranch Tests ----

func TestCreateFeatureBranch_NotARepo(t *testing.T) {
	h := newGitTestHandler(t)
	result, _, err := h.CreateFeatureBranch(context.Background(), &mcp.CallToolRequest{}, CreateFeatureBranchParams{
		RepoPath: "nonexistent-repo", BranchName: "feature/test",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Equal(t, "NOT_REPOSITORY", m["code"])
}

// ---- SyncWithRemote Tests ----

func TestSyncWithRemote_NotARepo(t *testing.T) {
	h := newGitTestHandler(t)
	result, _, err := h.SyncWithRemote(context.Background(), &mcp.CallToolRequest{}, SyncWithRemoteParams{
		RepoPath: "nonexistent-repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Equal(t, "NOT_REPOSITORY", m["code"])
}

// ---- CommitAndPushChanges Tests ----

func TestCommitAndPushChanges_NoWallet(t *testing.T) {
	h := newGitTestHandler(t)
	h.GClient = nil
	result, _, err := h.CommitAndPushChanges(context.Background(), &mcp.CallToolRequest{}, CommitAndPushChangesParams{
		RepoPath: "some-repo", CommitMessage: "test",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Contains(t, m["message"], "Authentication failed")
}

// ---- SearchCode Tests ----

func TestSearchCode_NotARepo(t *testing.T) {
	h := newGitTestHandler(t)
	result, _, err := h.SearchCode(context.Background(), &mcp.CallToolRequest{}, SearchCodeParams{
		RepoPath: "nonexistent-repo", Query: "test",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	m := parseResponse(t, result)
	assert.Equal(t, "NOT_REPOSITORY", m["code"])
}
