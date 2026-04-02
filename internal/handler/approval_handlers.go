package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ConfirmTransactionParams struct {
	PendingID string `json:"pending_id" jsonschema:"The ID of the pending transaction to confirm"`
}

// ConfirmTransaction broadcasts a previously held pending transaction.
func (h *ToolHandler) ConfirmTransaction(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ConfirmTransactionParams,
) (*mcp.CallToolResult, any, error) {
	if h.PendingTxs == nil {
		return toolErrorf(ErrApproval, "Approval mode is not enabled. No pending transactions to confirm.")
	}

	pending := h.PendingTxs.Get(p.PendingID)
	if pending == nil {
		return toolErrorf(ErrNotFound, "Pending transaction %q not found or expired.", p.PendingID)
	}

	// Verify session ownership
	var sid string
	if req != nil && req.Session != nil {
		sid = req.Session.ID()
	}
	if pending.SessionID != sid {
		return toolErrorf(ErrAuthFailed, "This pending transaction belongs to a different session.")
	}

	// Broadcast
	start := time.Now()
	_, err := pending.wallet.SignAndBroadcast(ctx, h.GClient.GetConn(), pending.msgs)
	AuditLog(pending.ToolName, pending.WalletAddress, err == nil, time.Since(start), pending.Detail)
	h.PendingTxs.Remove(p.PendingID)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to broadcast %s: %s", pending.ToolName, err)
	}

	return toolSuccess("Transaction confirmed and broadcast", map[string]any{
		"tool":   pending.ToolName,
		"detail": pending.Detail,
	})
}

type RejectTransactionParams struct {
	PendingID string `json:"pending_id" jsonschema:"The ID of the pending transaction to reject"`
}

// RejectTransaction cancels a previously held pending transaction.
func (h *ToolHandler) RejectTransaction(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p RejectTransactionParams,
) (*mcp.CallToolResult, any, error) {
	if h.PendingTxs == nil {
		return toolErrorf(ErrApproval, "Approval mode is not enabled. No pending transactions to reject.")
	}

	pending := h.PendingTxs.Get(p.PendingID)
	if pending == nil {
		return toolErrorf(ErrNotFound, "Pending transaction %q not found or expired.", p.PendingID)
	}

	// Verify session ownership
	var sid string
	if req != nil && req.Session != nil {
		sid = req.Session.ID()
	}
	if pending.SessionID != sid {
		return toolErrorf(ErrAuthFailed, "This pending transaction belongs to a different session.")
	}

	AuditLog(pending.ToolName, pending.WalletAddress, true, 0, fmt.Sprintf("rejected: %s", pending.Detail))
	h.PendingTxs.Remove(p.PendingID)

	return toolSuccess("Transaction rejected", map[string]any{
		"tool":   pending.ToolName,
		"detail": pending.Detail,
	})
}

// ListPendingTransactions lists all non-expired pending transactions for the current session.
func (h *ToolHandler) ListPendingTransactions(
	ctx context.Context,
	req *mcp.CallToolRequest,
	_ struct{},
) (*mcp.CallToolResult, any, error) {
	if h.PendingTxs == nil {
		return toolErrorf(ErrApproval, "Approval mode is not enabled.")
	}

	var sid string
	if req != nil && req.Session != nil {
		sid = req.Session.ID()
	}

	pending := h.PendingTxs.ListForSession(sid)
	if len(pending) == 0 {
		return toolSuccessData("No pending transactions", []any{})
	}

	items := make([]map[string]any, len(pending))
	for i, tx := range pending {
		items[i] = map[string]any{
			"pending_id": tx.ID,
			"tool":       tx.ToolName,
			"detail":     tx.Detail,
			"wallet":     tx.WalletAddress,
			"expires_at": tx.ExpiresAt.Format(time.RFC3339),
		}
	}
	return toolSuccessData(fmt.Sprintf("Found %d pending transactions", len(items)), items)
}
