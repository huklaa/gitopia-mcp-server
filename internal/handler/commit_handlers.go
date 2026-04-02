package handler

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListCommitsParams struct {
	Owner  string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name   string `json:"name" jsonschema:"Repository name"`
	Branch string `json:"branch,omitempty" jsonschema:"Branch name (defaults to repo's default branch)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of commits to return (default 50)"`
}

func (h *ToolHandler) ListCommits(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params ListCommitsParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}

	limit := params.Limit
	if limit == 0 {
		limit = 50
	}

	// Resolve owner/name to repository ID.
	repo, err := h.GClient.GetRepository(ctx, params.Owner, params.Name)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get repository '%s/%s': %s", params.Owner, params.Name, err)
	}

	// Use repo's default branch if caller didn't specify one.
	branch := params.Branch
	if branch == "" {
		branch = repo.DefaultBranch
		if branch == "" {
			branch = "main"
		}
	}

	// Resolve branch to HEAD SHA using direct branch query (avoids pagination).
	branchObj, err := h.GClient.GetBranch(ctx, params.Owner, params.Name, branch)
	if err != nil {
		return toolErrorf(ErrNotFound, "Branch '%s' not found in '%s/%s': %s", branch, params.Owner, params.Name, err)
	}

	commits, total, err := h.GClient.ListCommits(ctx, repo.Id, branchObj.Sha, limit)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to fetch commits for '%s/%s' branch '%s': %s", params.Owner, params.Name, branch, err)
	}

	return toolSuccessData(
		fmt.Sprintf("Found %d commits (of %d total) for '%s/%s' branch '%s'", len(commits), total, params.Owner, params.Name, branch),
		commits,
	)
}
