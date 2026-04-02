package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListReleasesParams struct {
	Owner string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name  string `json:"name" jsonschema:"Repository name"`
	Limit uint64 `json:"limit,omitempty" jsonschema:"Maximum number of releases to return (default 50)"`
}

func (h *ToolHandler) ListReleases(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params ListReleasesParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	limit := params.Limit
	if limit == 0 {
		limit = 50
	}
	releases, err := h.GClient.ListReleases(ctx, params.Owner, params.Name, limit)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list releases for '%s/%s': %s", params.Owner, params.Name, err)
	}
	out := make([]map[string]any, 0, len(releases))
	for _, r := range releases {
		out = append(out, map[string]any{
			"id":          r.Id,
			"tag_name":    r.TagName,
			"name":        r.Name,
			"description": r.Description,
			"draft":       r.Draft,
			"pre_release": r.PreRelease,
			"created_at":  r.CreatedAt,
			"updated_at":  r.UpdatedAt,
		})
	}
	return toolSuccessData(fmt.Sprintf("Found %d releases for '%s/%s'", len(out), params.Owner, params.Name), out)
}

type CreateReleaseParams struct {
	Owner       string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name        string `json:"name" jsonschema:"Repository name"`
	TagName     string `json:"tag_name" jsonschema:"Tag name for the release (e.g. v1.0.0)"`
	Target      string `json:"target,omitempty" jsonschema:"Branch or commit SHA to tag (defaults to default branch HEAD)"`
	ReleaseName string `json:"release_name" jsonschema:"Human-readable release title"`
	Description string `json:"description,omitempty" jsonschema:"Release notes / description"`
	Draft       bool   `json:"draft,omitempty" jsonschema:"Mark as draft release"`
	PreRelease  bool   `json:"pre_release,omitempty" jsonschema:"Mark as pre-release"`
	Provider    string `json:"provider,omitempty" jsonschema:"Git server provider address (defaults to gitopia15nv5vf6fmww8cxr6emrzxjvj36x5n8xvsxsqpw)"`
}

func (h *ToolHandler) CreateRelease(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreateReleaseParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("create_release", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	target := p.Target
	if target == "" {
		// Look up the repo's default branch instead of hardcoding "main"
		repo, err := h.GClient.GetRepository(ctx, p.Owner, p.Name)
		if err != nil {
			return toolErrorf(ErrNotFound, "Failed to get repository '%s/%s' for default branch: %s", p.Owner, p.Name, err)
		}
		target = repo.DefaultBranch
		if target == "" {
			target = "main"
		}
	}

	provider := p.Provider
	if provider == "" {
		provider = "gitopia15nv5vf6fmww8cxr6emrzxjvj36x5n8xvsxsqpw"
	}

	start := time.Now()
	res, err := h.GClient.CreateRelease(ctx, w, p.Owner, p.Name,
		p.TagName, target, p.ReleaseName, p.Description,
		p.Draft, p.PreRelease, provider)
	AuditLog("create_release", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to create release '%s' for '%s/%s': %s", p.TagName, p.Owner, p.Name, err)
	}

	return toolSuccess("Created release", map[string]any{"release_id": res.Id})
}
