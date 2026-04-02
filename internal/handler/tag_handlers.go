package handler

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListTagsParams struct {
	Owner string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name  string `json:"name" jsonschema:"Repository name"`
	Limit uint64 `json:"limit,omitempty" jsonschema:"Maximum number of tags to return (default 100)"`
}

func (h *ToolHandler) ListTags(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params ListTagsParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	limit := params.Limit
	if limit == 0 {
		limit = 100
	}
	tags, err := h.GClient.ListTags(ctx, params.Owner, params.Name, limit)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list tags for '%s/%s': %s", params.Owner, params.Name, err)
	}
	out := make([]map[string]any, 0, len(tags))
	for _, t := range tags {
		out = append(out, map[string]any{
			"name":       t.Name,
			"sha":        t.Sha,
			"created_at": t.CreatedAt,
			"updated_at": t.UpdatedAt,
		})
	}
	return toolSuccessData(fmt.Sprintf("Found %d tags for '%s/%s'", len(out), params.Owner, params.Name), out)
}
