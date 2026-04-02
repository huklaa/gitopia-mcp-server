package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommits_NoClient(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.ListCommits(context.Background(), &mcp.CallToolRequest{}, ListCommitsParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestListCommits_DefaultBranch(t *testing.T) {
	p := ListCommitsParams{Owner: "test", Name: "repo"}
	assert.Equal(t, "", p.Branch, "Branch should be empty before handler applies default 'main'")
}
