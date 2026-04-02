package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PendingTx represents a chain transaction awaiting user confirmation.
type PendingTx struct {
	ID            string    `json:"id"`
	ToolName      string    `json:"tool_name"`
	Detail        string    `json:"detail"`
	WalletAddress string    `json:"wallet_address"`
	SessionID     string    `json:"session_id"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	msgs          []sdk.Msg
	wallet        signing.Wallet
}

// PendingTxStore is a thread-safe store for pending transactions.
type PendingTxStore struct {
	mu    sync.RWMutex
	store map[string]*PendingTx
	ttl   time.Duration
}

// NewPendingTxStore creates a new store with the given TTL.
func NewPendingTxStore(ttl time.Duration) *PendingTxStore {
	return &PendingTxStore{
		store: make(map[string]*PendingTx),
		ttl:   ttl,
	}
}

// Add stores a pending transaction and returns its ID.
func (s *PendingTxStore) Add(tx *PendingTx) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	tx.ID = id
	tx.CreatedAt = time.Now()
	tx.ExpiresAt = time.Now().Add(s.ttl)

	s.mu.Lock()
	s.store[id] = tx
	s.mu.Unlock()
	return id
}

// Get returns a pending transaction by ID, or nil if missing or expired.
func (s *PendingTxStore) Get(id string) *PendingTx {
	s.mu.RLock()
	tx, ok := s.store[id]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	if time.Now().After(tx.ExpiresAt) {
		s.Remove(id)
		return nil
	}
	return tx
}

// Remove deletes a pending transaction by ID.
func (s *PendingTxStore) Remove(id string) {
	s.mu.Lock()
	delete(s.store, id)
	s.mu.Unlock()
}

// ListForSession returns all non-expired pending transactions for a session.
func (s *PendingTxStore) ListForSession(sessionID string) []*PendingTx {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	var result []*PendingTx
	for _, tx := range s.store {
		if tx.SessionID == sessionID && now.Before(tx.ExpiresAt) {
			result = append(result, tx)
		}
	}
	return result
}

// Cleanup removes all expired transactions.
func (s *PendingTxStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, tx := range s.store {
		if now.After(tx.ExpiresAt) {
			delete(s.store, id)
		}
	}
}

// signOrHold either broadcasts immediately or holds the transaction for approval.
// When approval mode is off (or PendingTxs is nil), it broadcasts directly via
// w.SignAndBroadcast. When approval mode is on, it stores the transaction and
// returns a preview result with the pending_id.
//
// Returns:
//   - (*CallToolResult, nil, nil) when held for approval
//   - (nil, txResp, err) when broadcast directly (caller handles result formatting)
func (h *ToolHandler) signOrHold(
	ctx context.Context,
	req *mcp.CallToolRequest,
	w signing.Wallet,
	msgs []sdk.Msg,
	toolName string,
	detail string,
) (*mcp.CallToolResult, any, error) {
	if !h.ApprovalMode || h.PendingTxs == nil {
		// Direct broadcast
		start := time.Now()
		txResp, err := w.SignAndBroadcast(ctx, h.GClient.GetConn(), msgs)
		AuditLog(toolName, w.Address(), err == nil, time.Since(start), detail)
		if err != nil {
			return toolErrorf(ErrChainTx, "Failed to execute %s: %s", toolName, err)
		}
		return nil, txResp, nil
	}

	// Hold for approval
	var sid string
	if req != nil && req.Session != nil {
		sid = req.Session.ID()
	}

	pendingID := h.PendingTxs.Add(&PendingTx{
		ToolName:      toolName,
		Detail:        detail,
		WalletAddress: w.Address(),
		SessionID:     sid,
		msgs:          msgs,
		wallet:        w,
	})

	preview := map[string]any{
		"status":            "success",
		"approval_required": true,
		"pending_id":        pendingID,
		"tool":              toolName,
		"detail":            detail,
		"wallet":            w.Address(),
		"expires_at":        h.PendingTxs.Get(pendingID).ExpiresAt.Format(time.RFC3339),
		"message":           "Use confirm_transaction with this pending_id to broadcast, or reject_transaction to cancel.",
	}
	data, err := json.Marshal(preview)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal approval preview: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}
