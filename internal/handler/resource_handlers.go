package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ResourceHandler holds dependencies for MCP resource handlers.
type ResourceHandler struct {
	GClient *gitopia.Client
}

// parseURIParams extracts named segments from a URI given a pattern.
// Pattern uses {name} placeholders, e.g. "gitopia://repos/{owner}/{name}".
func parseURIParams(uri, pattern string) (map[string]string, error) {
	patternParts := strings.Split(pattern, "/")
	uriParts := strings.Split(uri, "/")

	if len(uriParts) != len(patternParts) {
		return nil, fmt.Errorf("URI %q does not match pattern %q", uri, pattern)
	}

	params := make(map[string]string)
	for i, pp := range patternParts {
		if strings.HasPrefix(pp, "{") && strings.HasSuffix(pp, "}") {
			name := pp[1 : len(pp)-1]
			params[name] = uriParts[i]
		} else if pp != uriParts[i] {
			return nil, fmt.Errorf("URI %q does not match pattern %q at segment %d", uri, pattern, i)
		}
	}
	return params, nil
}

// HandleRepoResource reads a repository resource by URI.
func (rh *ResourceHandler) HandleRepoResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if rh.GClient == nil {
		return nil, fmt.Errorf("gitopia client is not available")
	}
	params, err := parseURIParams(req.Params.URI, "gitopia://repos/{owner}/{name}")
	if err != nil {
		return nil, err
	}
	repo, err := rh.GClient.GetRepository(context.Background(), params["owner"], params["name"])
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	out := map[string]any{
		"name":        repo.Name,
		"id":          repo.Id,
		"description": repo.Description,
		"stars":       repo.Stargazers,
		"forks":       repo.Forks,
		"owner":       repo.Owner.Id,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}

// HandleIssueResource reads an issue resource by URI.
func (rh *ResourceHandler) HandleIssueResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if rh.GClient == nil {
		return nil, fmt.Errorf("gitopia client is not available")
	}
	params, err := parseURIParams(req.Params.URI, "gitopia://repos/{owner}/{name}/issues/{iid}")
	if err != nil {
		return nil, err
	}
	iid, err := strconv.ParseUint(params["iid"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid issue IID: %w", err)
	}
	issue, err := rh.GClient.GetIssue(context.Background(), params["owner"], params["name"], iid)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	out := map[string]any{
		"number":      issue.Iid,
		"title":       issue.Title,
		"state":       issue.State.String(),
		"author":      issue.Creator,
		"description": issue.Description,
		"comments":    issue.CommentsCount,
		"labels":      issue.Labels,
		"assignees":   issue.Assignees,
		"created_at":  issue.CreatedAt,
		"updated_at":  issue.UpdatedAt,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}

// HandlePullRequestResource reads a pull request resource by URI.
func (rh *ResourceHandler) HandlePullRequestResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if rh.GClient == nil {
		return nil, fmt.Errorf("gitopia client is not available")
	}
	params, err := parseURIParams(req.Params.URI, "gitopia://repos/{owner}/{name}/pulls/{iid}")
	if err != nil {
		return nil, err
	}
	iid, err := strconv.ParseUint(params["iid"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid pull request IID: %w", err)
	}
	pr, err := rh.GClient.GetPullRequest(context.Background(), params["owner"], params["name"], iid)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}
	out := map[string]any{
		"iid":         pr.Iid,
		"title":       pr.Title,
		"state":       pr.State.String(),
		"description": pr.Description,
		"head":        pr.Head,
		"base":        pr.Base,
		"creator":     pr.Creator,
		"reviewers":   pr.Reviewers,
		"assignees":   pr.Assignees,
		"labels":      pr.Labels,
		"created_at":  pr.CreatedAt,
		"updated_at":  pr.UpdatedAt,
		"merged_by":   pr.MergedBy,
		"merged_at":   pr.MergedAt,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}

// HandleBountyResource reads a bounty resource by URI.
func (rh *ResourceHandler) HandleBountyResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if rh.GClient == nil {
		return nil, fmt.Errorf("gitopia client is not available")
	}
	params, err := parseURIParams(req.Params.URI, "gitopia://bounties/{id}")
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bounty ID: %w", err)
	}
	bounty, err := rh.GClient.GetBounty(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bounty: %w", err)
	}
	out := map[string]any{
		"id":            bounty.Id,
		"amount":        bounty.Amount,
		"state":         bounty.State.String(),
		"repository_id": bounty.RepositoryId,
		"parent_iid":    bounty.ParentIid,
		"parent":        bounty.Parent.String(),
		"expire_at":     bounty.ExpireAt,
		"rewarded_to":   bounty.RewardedTo,
		"creator":       bounty.Creator,
		"created_at":    bounty.CreatedAt,
		"updated_at":    bounty.UpdatedAt,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}
