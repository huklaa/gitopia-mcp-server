package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- CreateFeatureBranchPR Tests ----

func TestCreateFeatureBranchPR_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CreateFeatureBranchPR(context.Background(), &mcp.CallToolRequest{}, CreateFeatureBranchPRParams{
		RepoPath: "myrepo", Owner: "test", Name: "repo",
		BranchName: "feature/test", BaseBranch: "main",
		PRTitle: "Test PR", PRDescription: "Test",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestCreateFeatureBranchPR_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CreateFeatureBranchPR(context.Background(), &mcp.CallToolRequest{}, CreateFeatureBranchPRParams{
		RepoPath: "myrepo", Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- BootstrapRepo Tests ----

func TestBootstrapRepo_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.BootstrapRepo(context.Background(), &mcp.CallToolRequest{}, BootstrapRepoParams{
		OwnerId: "user", Name: "repo", Description: "test",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestBootstrapRepo_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.BootstrapRepo(context.Background(), &mcp.CallToolRequest{}, BootstrapRepoParams{
		OwnerId: "user", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

