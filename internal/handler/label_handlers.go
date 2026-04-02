package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListLabelsParams struct {
	Owner string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name  string `json:"name" jsonschema:"Repository name"`
}

func (h *ToolHandler) ListLabels(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params ListLabelsParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	repo, err := h.GClient.GetRepository(ctx, params.Owner, params.Name)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get repository '%s/%s': %s", params.Owner, params.Name, err)
	}
	out := make([]map[string]any, 0, len(repo.Labels))
	for _, l := range repo.Labels {
		out = append(out, map[string]any{
			"id":          l.Id,
			"name":        l.Name,
			"color":       l.Color,
			"description": l.Description,
		})
	}
	return toolSuccessData(fmt.Sprintf("Found %d labels for '%s/%s'", len(out), params.Owner, params.Name), out)
}

type CreateLabelParams struct {
	Owner       string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name        string `json:"name" jsonschema:"Repository name"`
	LabelName   string `json:"label_name" jsonschema:"Label name (3-63 characters)"`
	Color       string `json:"color" jsonschema:"Label color as hex code (e.g. FF0000)"`
	Description string `json:"description,omitempty" jsonschema:"Label description (max 255 characters)"`
}

func (h *ToolHandler) CreateLabel(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreateLabelParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("create_label", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	start := time.Now()
	res, err := h.GClient.CreateRepositoryLabel(ctx, w, p.Owner, p.Name, p.LabelName, p.Color, p.Description)
	AuditLog("create_label", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to create label '%s' for '%s/%s': %s", p.LabelName, p.Owner, p.Name, err)
	}

	return toolSuccess("Created label", map[string]any{"label_id": res.Id})
}

type DeleteLabelParams struct {
	Owner   string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name    string `json:"name" jsonschema:"Repository name"`
	LabelID uint64 `json:"label_id" jsonschema:"Label ID to delete (from list_labels)"`
}

func (h *ToolHandler) DeleteLabel(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p DeleteLabelParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("delete_label", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	start := time.Now()
	_, err = h.GClient.DeleteRepositoryLabel(ctx, w, p.Owner, p.Name, p.LabelID)
	AuditLog("delete_label", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to delete label %d for '%s/%s': %s", p.LabelID, p.Owner, p.Name, err)
	}

	return toolSuccess(fmt.Sprintf("Deleted label %d from '%s/%s'", p.LabelID, p.Owner, p.Name), nil)
}
