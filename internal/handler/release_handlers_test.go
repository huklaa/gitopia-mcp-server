package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListReleases_NoClient(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.ListReleases(context.Background(), &mcp.CallToolRequest{}, ListReleasesParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestCreateRelease_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CreateRelease(context.Background(), &mcp.CallToolRequest{}, CreateReleaseParams{
		Owner: "test", Name: "repo", TagName: "v1.0.0", ReleaseName: "First Release",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestCreateRelease_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CreateRelease(context.Background(), &mcp.CallToolRequest{}, CreateReleaseParams{
		Owner: "test", Name: "repo", TagName: "v1.0.0", ReleaseName: "First Release",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
