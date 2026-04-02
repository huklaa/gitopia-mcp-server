package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- GetPullRequest Tests ----

func TestGetPullRequest_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.GetPullRequest(context.Background(), &mcp.CallToolRequest{}, GetPullRequestParams{
		Owner: "test", Name: "repo", PullIid: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- GetPullRequestDiff Tests ----

func TestGetPullRequestDiff_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.GetPullRequestDiff(context.Background(), &mcp.CallToolRequest{}, GetPullRequestDiffParams{
		Owner: "test", Name: "repo", PullIid: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- CommentOnPullRequest Tests ----

func TestCommentOnPullRequest_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0) // zero limit
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CommentOnPullRequest(context.Background(), &mcp.CallToolRequest{}, CommentOnPullRequestParams{
		Owner: "test", Name: "repo", PullIid: 1, Body: "comment",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestCommentOnPullRequest_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CommentOnPullRequest(context.Background(), &mcp.CallToolRequest{}, CommentOnPullRequestParams{
		Owner: "test", Name: "repo", PullIid: 1, Body: "comment",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

// ---- ForkRepository Tests ----

func TestForkRepository_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.ForkRepository(context.Background(), &mcp.CallToolRequest{}, ForkRepositoryParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestForkRepository_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.ForkRepository(context.Background(), &mcp.CallToolRequest{}, ForkRepositoryParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- GetIssue Tests ----

func TestGetIssue_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.GetIssue(context.Background(), &mcp.CallToolRequest{}, GetIssueParams{
		Owner: "test", Name: "repo", IssueIid: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- CommentOnIssue Tests ----

func TestCommentOnIssue_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CommentOnIssue(context.Background(), &mcp.CallToolRequest{}, CommentOnIssueParams{
		Owner: "test", Name: "repo", IssueIid: 1, Body: "comment",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestCommentOnIssue_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CommentOnIssue(context.Background(), &mcp.CallToolRequest{}, CommentOnIssueParams{
		Owner: "test", Name: "repo", IssueIid: 1, Body: "comment",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

// ---- UpdateIssue Tests ----

func TestUpdateIssue_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.UpdateIssue(context.Background(), &mcp.CallToolRequest{}, UpdateIssueParams{
		Owner: "test", Name: "repo", IssueIid: 1, ToggleState: true,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestUpdateIssue_NoActions(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.UpdateIssue(context.Background(), &mcp.CallToolRequest{}, UpdateIssueParams{
		Owner: "test", Name: "repo", IssueIid: 1,
		// No actions specified
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "No update actions specified")
}

func TestUpdateIssue_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.UpdateIssue(context.Background(), &mcp.CallToolRequest{}, UpdateIssueParams{
		Owner: "test", Name: "repo", IssueIid: 1, ToggleState: true,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "rate limit")
}

// ---- CreateIssue Tests ----

func TestCreateIssue_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CreateIssue(context.Background(), &mcp.CallToolRequest{}, CreateIssueParams{
		Owner: "test", Name: "repo", Title: "Bug", Description: "desc",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestCreateIssue_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CreateIssue(context.Background(), &mcp.CallToolRequest{}, CreateIssueParams{
		Owner: "test", Name: "repo", Title: "Bug", Description: "desc",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- ListRepos Tests ----

func TestListRepos_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.ListRepos(context.Background(), &mcp.CallToolRequest{}, ListReposParams{
		Owner: "test",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- GetRepo Tests ----

func TestGetRepo_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.GetRepo(context.Background(), &mcp.CallToolRequest{}, GetRepoParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- ListBranches Tests ----

func TestListBranches_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.ListBranches(context.Background(), &mcp.CallToolRequest{}, ListBranchesParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- GetFile Tests ----

func TestGetFile_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.GetFile(context.Background(), &mcp.CallToolRequest{}, GetFileParams{
		Owner: "test", Name: "repo", Branch: "main", Path: "README.md",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- ListIssues Tests ----

func TestListIssues_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.ListIssues(context.Background(), &mcp.CallToolRequest{}, ListIssuesParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- ListPullRequests Tests ----

func TestListPullRequests_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.ListPullRequests(context.Background(), &mcp.CallToolRequest{}, ListPullRequestsParams{
		Owner: "test", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestListPullRequests_DefaultLimit(t *testing.T) {
	p := ListPullRequestsParams{}
	assert.Equal(t, uint64(0), p.Limit)
}

// ---- CreateRepository Tests ----

func TestCreateRepository_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CreateRepository(context.Background(), &mcp.CallToolRequest{}, CreateRepoParams{
		OwnerId: "user", Name: "repo", Description: "test",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestCreateRepository_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CreateRepository(context.Background(), &mcp.CallToolRequest{}, CreateRepoParams{
		OwnerId: "user", Name: "repo",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- MergePullRequest Tests ----

func TestMergePullRequest_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.MergePullRequest(context.Background(), &mcp.CallToolRequest{}, MergePullRequestParams{
		Owner: "test", Name: "repo", PullIid: 1,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

// ---- CreatePullRequest Tests ----

func TestCreatePullRequest_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CreatePullRequest(context.Background(), &mcp.CallToolRequest{}, CreatePullRequestParams{
		Owner: "test", Name: "repo", Title: "PR", Description: "desc",
		HeadBranch: "feature", BaseBranch: "main",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

