package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- CreateBounty Tests ----

func TestCreateBounty_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CreateBounty(context.Background(), &mcp.CallToolRequest{}, CreateBountyParams{
		Owner: "test", Name: "repo", IssueIID: 1,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestCreateBounty_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CreateBounty(context.Background(), &mcp.CallToolRequest{}, CreateBountyParams{
		Owner: "test", Name: "repo", IssueIID: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- UpdateBounty Tests ----

func TestUpdateBounty_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.UpdateBounty(context.Background(), &mcp.CallToolRequest{}, UpdateBountyParams{
		BountyID: 1, Expiry: 1000,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestUpdateBounty_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.UpdateBounty(context.Background(), &mcp.CallToolRequest{}, UpdateBountyParams{
		BountyID: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- CloseBounty Tests ----

func TestCloseBounty_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CloseBounty(context.Background(), &mcp.CallToolRequest{}, CloseBountyParams{
		BountyID: 1,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestCloseBounty_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CloseBounty(context.Background(), &mcp.CallToolRequest{}, CloseBountyParams{
		BountyID: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- DeleteBounty Tests ----

func TestDeleteBounty_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.DeleteBounty(context.Background(), &mcp.CallToolRequest{}, DeleteBountyParams{
		BountyID: 1,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestDeleteBounty_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.DeleteBounty(context.Background(), &mcp.CallToolRequest{}, DeleteBountyParams{
		BountyID: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- ListBounties Tests ----

func TestListBounties_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.ListBounties(context.Background(), &mcp.CallToolRequest{}, ListBountiesParams{})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestListBounties_DefaultLimit(t *testing.T) {
	p := ListBountiesParams{}
	assert.Equal(t, uint64(0), p.Limit)
}

// ---- GetBounty Tests ----

func TestGetBounty_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.GetBounty(context.Background(), &mcp.CallToolRequest{}, GetBountyParams{BountyID: 1})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}
