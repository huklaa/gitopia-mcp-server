package handler

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CoinParam is a JSON-schema-friendly representation of a coin amount.
// sdk.Coin uses math.Int which produces "type":"object" in JSON schema;
// this struct produces correct "type":"string" for the amount field.
type CoinParam struct {
	Denom  string `json:"denom" jsonschema:"Token denomination (e.g. ulore)"`
	Amount string `json:"amount" jsonschema:"Token amount as string (e.g. 1000)"`
}

type CreateBountyParams struct {
	Owner    string      `json:"owner"`
	Name     string      `json:"name"`
	IssueIID uint64      `json:"issue_iid"`
	Amount   []CoinParam `json:"amount"`
	Expiry   int64       `json:"expiry,omitempty"`
}

func (h *ToolHandler) CreateBounty(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreateBountyParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("create_bounty", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	repo, err := h.GClient.GetRepository(ctx, p.Owner, p.Name)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get repository '%s/%s': %s", p.Owner, p.Name, err)
	}

	coins := make([]sdk.Coin, len(p.Amount))
	for i, c := range p.Amount {
		amt, ok := sdk.NewIntFromString(c.Amount)
		if !ok {
			return toolErrorf(ErrValidation, "Invalid amount '%s' for denom '%s'", c.Amount, c.Denom)
		}
		coins[i] = sdk.NewCoin(c.Denom, amt)
	}

	msg := &gitopiatypes.MsgCreateBounty{
		Creator:      w.Address(),
		RepositoryId: repo.Id,
		ParentIid:    p.IssueIID,
		Parent:       gitopiatypes.BountyParentIssue,
		Amount:       coins,
		Expiry:       p.Expiry,
	}

	start := time.Now()
	res, err := h.GClient.CreateBounty(ctx, w, msg)
	AuditLog("create_bounty", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to create bounty: %s", err)
	}

	return toolSuccess("Created bounty", map[string]any{"bounty_id": res.Id})
}

type UpdateBountyParams struct {
	BountyID uint64 `json:"bounty_id"`
	Expiry   int64  `json:"expiry,omitempty"`
}

func (h *ToolHandler) UpdateBounty(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p UpdateBountyParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("update_bounty", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	msg := &gitopiatypes.MsgUpdateBountyExpiry{
		Creator: w.Address(),
		Id:      p.BountyID,
		Expiry:  p.Expiry,
	}

	start := time.Now()
	_, err = h.GClient.UpdateBounty(ctx, w, msg)
	AuditLog("update_bounty", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to update bounty %d: %s", p.BountyID, err)
	}

	return toolSuccess("Updated bounty expiry", map[string]any{"bounty_id": p.BountyID})
}

type CloseBountyParams struct {
	BountyID uint64 `json:"bounty_id"`
}

func (h *ToolHandler) CloseBounty(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CloseBountyParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("close_bounty", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	msg := &gitopiatypes.MsgCloseBounty{
		Creator: w.Address(),
		Id:      p.BountyID,
	}

	start := time.Now()
	_, err = h.GClient.CloseBounty(ctx, w, msg)
	AuditLog("close_bounty", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to close bounty %d: %s", p.BountyID, err)
	}

	return toolSuccess("Closed bounty", map[string]any{"bounty_id": p.BountyID})
}

type DeleteBountyParams struct {
	BountyID uint64 `json:"bounty_id"`
}

func (h *ToolHandler) DeleteBounty(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p DeleteBountyParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("delete_bounty", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	msg := &gitopiatypes.MsgDeleteBounty{
		Creator: w.Address(),
		Id:      p.BountyID,
	}

	start := time.Now()
	_, err = h.GClient.DeleteBounty(ctx, w, msg)
	AuditLog("delete_bounty", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to delete bounty %d: %s", p.BountyID, err)
	}

	return toolSuccess("Deleted bounty", map[string]any{"bounty_id": p.BountyID})
}

// ---- Bounty Query Handlers ----

type ListBountiesParams struct {
	Limit uint64 `json:"limit,omitempty" jsonschema:"Maximum number of bounties to return (default 50)"`
}

func (h *ToolHandler) ListBounties(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ListBountiesParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	limit := p.Limit
	if limit == 0 {
		limit = 50
	}
	bounties, err := h.GClient.ListBounties(ctx, limit)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list bounties: %s", err)
	}
	out := make([]map[string]any, 0, len(bounties))
	for _, b := range bounties {
		out = append(out, map[string]any{
			"id":            b.Id,
			"amount":        b.Amount,
			"state":         b.State.String(),
			"repository_id": b.RepositoryId,
			"parent_iid":    b.ParentIid,
			"expire_at":     b.ExpireAt,
			"creator":       b.Creator,
		})
	}
	return toolSuccessData(fmt.Sprintf("Found %d bounties", len(out)), out)
}

type GetBountyParams struct {
	BountyID uint64 `json:"bounty_id" jsonschema:"The ID of the bounty to retrieve"`
}

func (h *ToolHandler) GetBounty(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p GetBountyParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	bounty, err := h.GClient.GetBounty(ctx, p.BountyID)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get bounty %d: %s", p.BountyID, err)
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
	return toolSuccessData(fmt.Sprintf("Bounty %d", p.BountyID), out)
}
