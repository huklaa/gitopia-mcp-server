package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixIssuePrompt_ValidArgs(t *testing.T) {
	ph := &PromptHandler{}
	result, err := ph.FixIssuePrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"owner":        "myorg",
				"repo":         "myrepo",
				"issue_number": "42",
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.Description, "#42")
	assert.Contains(t, result.Description, "myorg/myrepo")
	require.Len(t, result.Messages, 1)
	assert.Equal(t, mcp.Role("user"), result.Messages[0].Role)
	text := result.Messages[0].Content.(*mcp.TextContent).Text
	assert.Contains(t, text, "get_issue")
	assert.Contains(t, text, "git_clone")
	assert.Contains(t, text, "create_pull_request")
}

func TestFixIssuePrompt_MissingOwner(t *testing.T) {
	ph := &PromptHandler{}
	_, err := ph.FixIssuePrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"repo":         "myrepo",
				"issue_number": "42",
			},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required arguments")
}

func TestFixIssuePrompt_MissingRepo(t *testing.T) {
	ph := &PromptHandler{}
	_, err := ph.FixIssuePrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"owner":        "myorg",
				"issue_number": "42",
			},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required arguments")
}

func TestFixIssuePrompt_MissingIssueNumber(t *testing.T) {
	ph := &PromptHandler{}
	_, err := ph.FixIssuePrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"owner": "myorg",
				"repo":  "myrepo",
			},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required arguments")
}

func TestReviewPRPrompt_ValidArgs(t *testing.T) {
	ph := &PromptHandler{}
	result, err := ph.ReviewPRPrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"owner":     "myorg",
				"repo":      "myrepo",
				"pr_number": "15",
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.Description, "#15")
	require.Len(t, result.Messages, 1)
	text := result.Messages[0].Content.(*mcp.TextContent).Text
	assert.Contains(t, text, "get_pull_request")
	assert.Contains(t, text, "get_pull_request_diff")
	assert.Contains(t, text, "comment_on_pull_request")
}

func TestReviewPRPrompt_MissingArgs(t *testing.T) {
	ph := &PromptHandler{}
	_, err := ph.ReviewPRPrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"owner": "myorg",
			},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required arguments")
}

func TestHuntBountyPrompt_NoArgs(t *testing.T) {
	ph := &PromptHandler{}
	result, err := ph.HuntBountyPrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.Description, "bounties")
	require.Len(t, result.Messages, 1)
	text := result.Messages[0].Content.(*mcp.TextContent).Text
	assert.Contains(t, text, "list_bounties")
	assert.Contains(t, text, "get_bounty")
}

func TestHuntBountyPrompt_WithFilters(t *testing.T) {
	ph := &PromptHandler{}
	result, err := ph.HuntBountyPrompt(context.Background(), &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"min_amount": "1000",
				"state":      "active",
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	text := result.Messages[0].Content.(*mcp.TextContent).Text
	assert.Contains(t, text, "1000")
	assert.Contains(t, text, "active")
}
