package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTrustLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected TrustLevel
	}{
		{"readonly", TrustReadOnly},
		{"read_only", TrustReadOnly},
		{"ro", TrustReadOnly},
		{"READONLY", TrustReadOnly},
		{"localwrite", TrustLocalWrite},
		{"local_write", TrustLocalWrite},
		{"lw", TrustLocalWrite},
		{"chainwrite", TrustChainWrite},
		{"chain_write", TrustChainWrite},
		{"cw", TrustChainWrite},
		{"", TrustChainWrite},
		{"unknown", TrustChainWrite},
		{"  readonly  ", TrustReadOnly},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseTrustLevel(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTrustLevelString(t *testing.T) {
	assert.Equal(t, "readonly", TrustReadOnly.String())
	assert.Equal(t, "localwrite", TrustLocalWrite.String())
	assert.Equal(t, "chainwrite", TrustChainWrite.String())
	assert.Equal(t, "unknown", TrustLevel(99).String())
}

func TestCheckTrust_ReadOnlyLevel(t *testing.T) {
	// Read-only tools should work at readonly level
	readOnlyTools := []string{
		"list_repos", "get_repo", "list_branches", "get_file_contents",
		"list_issues", "get_issue", "list_pull_requests",
		"list_bounties", "get_bounty",
		"get_user_context", "set_active_dao", "refresh_user_context",
	}
	for _, tool := range readOnlyTools {
		t.Run("allow_"+tool, func(t *testing.T) {
			assert.NoError(t, CheckTrust(TrustReadOnly, tool))
		})
	}

	// Local-write tools should be blocked at readonly level
	localWriteTools := []string{
		"git_clone", "create_feature_branch", "sync_with_remote",
	}
	for _, tool := range localWriteTools {
		t.Run("block_"+tool, func(t *testing.T) {
			err := CheckTrust(TrustReadOnly, tool)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "trust level")
		})
	}

	// Chain-write tools should be blocked at readonly level
	chainWriteTools := []string{
		"create_repo", "create_issue", "create_pull_request",
		"merge_pull_request", "create_bounty", "create_dao",
	}
	for _, tool := range chainWriteTools {
		t.Run("block_"+tool, func(t *testing.T) {
			assert.Error(t, CheckTrust(TrustReadOnly, tool))
		})
	}
}

func TestCheckTrust_LocalWriteLevel(t *testing.T) {
	// Local-write tools should work at localwrite level
	assert.NoError(t, CheckTrust(TrustLocalWrite, "git_clone"))
	assert.NoError(t, CheckTrust(TrustLocalWrite, "create_feature_branch"))
	assert.NoError(t, CheckTrust(TrustLocalWrite, "sync_with_remote"))

	// Read-only tools should also work
	assert.NoError(t, CheckTrust(TrustLocalWrite, "list_repos"))

	// Chain-write tools should be blocked
	assert.Error(t, CheckTrust(TrustLocalWrite, "create_repo"))
	assert.Error(t, CheckTrust(TrustLocalWrite, "create_bounty"))
}

func TestCheckTrust_ChainWriteLevel(t *testing.T) {
	// All tools should work at chainwrite level
	assert.NoError(t, CheckTrust(TrustChainWrite, "list_repos"))
	assert.NoError(t, CheckTrust(TrustChainWrite, "git_clone"))
	assert.NoError(t, CheckTrust(TrustChainWrite, "create_repo"))
	assert.NoError(t, CheckTrust(TrustChainWrite, "create_bounty"))
}

func TestCheckTrust_UnknownTool(t *testing.T) {
	// Unknown tools default to chain-write requirement
	assert.NoError(t, CheckTrust(TrustChainWrite, "unknown_tool"))
	assert.Error(t, CheckTrust(TrustReadOnly, "unknown_tool"))
	assert.Error(t, CheckTrust(TrustLocalWrite, "unknown_tool"))
}

// TestWithTrust_BlocksExecution verifies that the withTrust wrapper actually
// prevents handler execution when trust level is insufficient.
func TestWithTrust_BlocksExecution(t *testing.T) {
	handlerCalled := false
	inner := func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		handlerCalled = true
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
		}, nil, nil
	}

	h := &ToolHandler{TrustLvl: TrustReadOnly}
	wrapped := withTrust(h, "create_issue", inner)

	result, _, err := wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})
	require.NoError(t, err)
	assert.False(t, handlerCalled, "handler should NOT be called when trust is insufficient")
	assert.True(t, result.IsError, "result should be marked as error")

	// Verify the error message contains access denied
	text := result.Content[0].(*mcp.TextContent).Text
	var resp map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &resp))
	assert.Equal(t, "error", resp["status"])
	assert.Contains(t, resp["message"], "Access denied")
}

// TestWithTrust_AllowsExecution verifies that the withTrust wrapper allows
// handler execution when trust level is sufficient.
func TestWithTrust_AllowsExecution(t *testing.T) {
	handlerCalled := false
	inner := func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		handlerCalled = true
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
		}, nil, nil
	}

	h := &ToolHandler{TrustLvl: TrustChainWrite}
	wrapped := withTrust(h, "create_issue", inner)

	result, _, err := wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})
	require.NoError(t, err)
	assert.True(t, handlerCalled, "handler SHOULD be called when trust is sufficient")
	assert.False(t, result.IsError)
}

// TestWithTrust_ReadOnlyAllowsReadTools verifies readonly trust allows read tools.
func TestWithTrust_ReadOnlyAllowsReadTools(t *testing.T) {
	handlerCalled := false
	inner := func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		handlerCalled = true
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
		}, nil, nil
	}

	h := &ToolHandler{TrustLvl: TrustReadOnly}
	wrapped := withTrust(h, "list_repos", inner)

	_, _, err := wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})
	require.NoError(t, err)
	assert.True(t, handlerCalled, "readonly trust should allow read tools")
}

// TestWithTrust_LocalWriteBlocksChainWrite verifies localwrite trust blocks chain-write tools.
func TestWithTrust_LocalWriteBlocksChainWrite(t *testing.T) {
	handlerCalled := false
	inner := func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		handlerCalled = true
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
		}, nil, nil
	}

	h := &ToolHandler{TrustLvl: TrustLocalWrite}
	wrapped := withTrust(h, "create_bounty", inner)

	result, _, err := wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})
	require.NoError(t, err)
	assert.False(t, handlerCalled, "localwrite trust should block chain-write tools")
	assert.True(t, result.IsError)
}
