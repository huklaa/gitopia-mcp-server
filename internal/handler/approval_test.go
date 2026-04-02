package handler

import (
	"context"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosgroup "github.com/cosmos/cosmos-sdk/x/group"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPendingTxStore_AddAndGet(t *testing.T) {
	store := NewPendingTxStore(5 * time.Minute)
	tx := &PendingTx{ToolName: "dao_vote", WalletAddress: "gitopia1abc"}
	id := store.Add(tx)
	if id == "" {
		t.Fatal("expected non-empty ID")
	}
	got := store.Get(id)
	if got == nil {
		t.Fatal("expected to find pending tx")
	}
	if got.ToolName != "dao_vote" {
		t.Fatalf("expected tool_name dao_vote, got %s", got.ToolName)
	}
}

func TestPendingTxStore_GetExpired(t *testing.T) {
	store := NewPendingTxStore(0) // 0 TTL = immediately expired
	tx := &PendingTx{ToolName: "dao_vote"}
	id := store.Add(tx)
	time.Sleep(time.Millisecond) // ensure expiry
	got := store.Get(id)
	if got != nil {
		t.Fatal("expected nil for expired tx")
	}
}

func TestPendingTxStore_Remove(t *testing.T) {
	store := NewPendingTxStore(5 * time.Minute)
	tx := &PendingTx{ToolName: "dao_vote"}
	id := store.Add(tx)
	store.Remove(id)
	if got := store.Get(id); got != nil {
		t.Fatal("expected nil after remove")
	}
}

func TestPendingTxStore_ListForSession(t *testing.T) {
	store := NewPendingTxStore(5 * time.Minute)
	store.Add(&PendingTx{ToolName: "t1", SessionID: "session-a"})
	store.Add(&PendingTx{ToolName: "t2", SessionID: "session-b"})
	store.Add(&PendingTx{ToolName: "t3", SessionID: "session-a"})

	listA := store.ListForSession("session-a")
	if len(listA) != 2 {
		t.Fatalf("expected 2 txs for session-a, got %d", len(listA))
	}
	listB := store.ListForSession("session-b")
	if len(listB) != 1 {
		t.Fatalf("expected 1 tx for session-b, got %d", len(listB))
	}
}

func TestPendingTxStore_Cleanup(t *testing.T) {
	store := NewPendingTxStore(time.Millisecond)

	// Add an expired tx
	store.Add(&PendingTx{ToolName: "expired"})
	time.Sleep(2 * time.Millisecond)

	// Add a live tx (manually adjust TTL for this one)
	store2 := NewPendingTxStore(5 * time.Minute)
	liveID := store2.Add(&PendingTx{ToolName: "live"})

	// Copy the live tx into the first store
	store.mu.Lock()
	store.store[liveID] = store2.store[liveID]
	store.mu.Unlock()

	store.Cleanup()

	store.mu.RLock()
	count := len(store.store)
	store.mu.RUnlock()
	if count != 1 {
		t.Fatalf("expected 1 tx after cleanup, got %d", count)
	}
}

func TestConfirmTransaction_NotEnabled(t *testing.T) {
	h := &ToolHandler{} // PendingTxs is nil
	result, _, err := h.ConfirmTransaction(context.Background(), &mcp.CallToolRequest{}, ConfirmTransactionParams{PendingID: "abc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result when approval mode is not enabled")
	}
}

func TestConfirmTransaction_NotFound(t *testing.T) {
	h := &ToolHandler{
		PendingTxs: NewPendingTxStore(5 * time.Minute),
	}
	result, _, err := h.ConfirmTransaction(context.Background(), &mcp.CallToolRequest{}, ConfirmTransactionParams{PendingID: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result for nonexistent pending tx")
	}
}

func TestRejectTransaction_Success(t *testing.T) {
	store := NewPendingTxStore(5 * time.Minute)
	id := store.Add(&PendingTx{ToolName: "dao_vote", WalletAddress: "gitopia1abc", SessionID: ""})

	h := &ToolHandler{PendingTxs: store}
	result, _, err := h.RejectTransaction(context.Background(), &mcp.CallToolRequest{}, RejectTransactionParams{PendingID: id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatal("expected success result")
	}
	if store.Get(id) != nil {
		t.Fatal("expected tx to be removed after rejection")
	}
}

func TestListPendingTransactions_Empty(t *testing.T) {
	h := &ToolHandler{PendingTxs: NewPendingTxStore(5 * time.Minute)}
	result, _, err := h.ListPendingTransactions(context.Background(), &mcp.CallToolRequest{}, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "No pending transactions") {
		t.Fatalf("unexpected text: %s", text)
	}
}

func TestSignOrHold_ApprovalMode(t *testing.T) {
	store := NewPendingTxStore(5 * time.Minute)
	h := &ToolHandler{
		ApprovalMode: true,
		PendingTxs:   store,
	}
	w := signing.NewTestWallet("gitopia1test", nil)
	msgs := []sdk.Msg{&cosmosgroup.MsgVote{
		ProposalId: 1,
		Voter:      "gitopia1test",
		Option:     cosmosgroup.VOTE_OPTION_YES,
	}}

	result, _, err := h.signOrHold(context.Background(), &mcp.CallToolRequest{}, w, msgs, "dao_vote", "proposal=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result (approval preview)")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "pending_id") {
		t.Fatalf("expected pending_id in result, got: %s", text)
	}
	if !strings.Contains(text, "approval_required") {
		t.Fatalf("expected approval_required in result, got: %s", text)
	}
}

func TestSignOrHold_DirectBroadcast(t *testing.T) {
	h := &ToolHandler{
		ApprovalMode: false,
		GClient:      &gitopia.Client{}, // zero-value client; testWallet ignores conn
	}
	w := signing.NewTestWallet("gitopia1test", nil)
	msgs := []sdk.Msg{&cosmosgroup.MsgVote{
		ProposalId: 1,
		Voter:      "gitopia1test",
		Option:     cosmosgroup.VOTE_OPTION_YES,
	}}

	result, txResp, err := h.signOrHold(context.Background(), &mcp.CallToolRequest{}, w, msgs, "dao_vote", "proposal=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result for direct broadcast")
	}
	if txResp == nil {
		t.Fatal("expected non-nil txResp for direct broadcast")
	}
}
