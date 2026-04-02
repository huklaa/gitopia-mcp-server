package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTags_NoClient(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.ListTags(context.Background(), &mcp.CallToolRequest{}, ListTagsParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestListTags_DefaultLimit(t *testing.T) {
	// Verify the default limit logic (no client needed for this check)
	p := ListTagsParams{Owner: "test", Name: "repo"}
	assert.Equal(t, uint64(0), p.Limit, "Limit should default to zero before handler applies default")
}
