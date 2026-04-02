package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListLabels_NoClient(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.ListLabels(context.Background(), &mcp.CallToolRequest{}, ListLabelsParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestCreateLabel_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CreateLabel(context.Background(), &mcp.CallToolRequest{}, CreateLabelParams{
		Owner: "test", Name: "repo", LabelName: "bug", Color: "FF0000",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestCreateLabel_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CreateLabel(context.Background(), &mcp.CallToolRequest{}, CreateLabelParams{
		Owner: "test", Name: "repo", LabelName: "bug", Color: "FF0000",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestDeleteLabel_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.DeleteLabel(context.Background(), &mcp.CallToolRequest{}, DeleteLabelParams{
		Owner: "test", Name: "repo", LabelID: 1,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestDeleteLabel_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.DeleteLabel(context.Background(), &mcp.CallToolRequest{}, DeleteLabelParams{
		Owner: "test", Name: "repo", LabelID: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
