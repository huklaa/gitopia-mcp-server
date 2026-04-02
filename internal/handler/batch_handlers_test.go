package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchExecute_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.BatchExecute(context.Background(), &mcp.CallToolRequest{}, BatchExecuteParams{
		Operations: []BatchOperation{
			{Tool: "dao_vote", Params: map[string]any{"proposal_id": float64(1), "option": "yes"}},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestBatchExecute_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.BatchExecute(context.Background(), &mcp.CallToolRequest{}, BatchExecuteParams{
		Operations: []BatchOperation{
			{Tool: "dao_vote", Params: map[string]any{"proposal_id": float64(1), "option": "yes"}},
		},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestBatchExecute_EmptyOperations(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.BatchExecute(context.Background(), &mcp.CallToolRequest{}, BatchExecuteParams{
		Operations: nil,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "at least one operation")
}

func TestBatchExecute_TooManyOperations(t *testing.T) {
	h := &ToolHandler{}
	ops := make([]BatchOperation, 11)
	for i := range ops {
		ops[i] = BatchOperation{Tool: "dao_vote", Params: map[string]any{"proposal_id": float64(1), "option": "yes"}}
	}
	result, _, err := h.BatchExecute(context.Background(), &mcp.CallToolRequest{}, BatchExecuteParams{
		Operations: ops,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "maximum 10")
}

func TestBuildSingleMessage_UnsupportedTool(t *testing.T) {
	op := BatchOperation{Tool: "create_repo", Params: map[string]any{"name": "test"}}
	_, err := buildSingleMessage(context.Background(), nil, op)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported in batch")
}

func TestBuildSingleMessage_InvalidParams(t *testing.T) {
	// Pass nil params to trigger marshal error on unmarshal
	op := BatchOperation{Tool: "dao_vote", Params: nil}
	_, err := buildSingleMessage(context.Background(), nil, op)
	// nil params will marshal to "null" which will fail to unmarshal into struct
	// The vote option "" will trigger "invalid vote option"
	require.Error(t, err)
}

func TestBuildSingleMessage_InvalidVoteOption(t *testing.T) {
	op := BatchOperation{
		Tool:   "dao_vote",
		Params: map[string]any{"proposal_id": float64(1), "option": "maybe"},
	}
	// nil wallet is fine here because parseVoteOption runs before w.Address()
	_, err := buildSingleMessage(context.Background(), nil, op)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid vote option")
}
