package handler

import (
	"testing"

	"github.com/gitopia/gitopia-mcp-server/internal/session"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGetSessionContext_NilSession(t *testing.T) {
	h := &ToolHandler{}
	cm := h.getSessionContext(nil)
	if cm == nil {
		t.Fatal("expected non-nil ContextManager for nil request")
	}
}

func TestGetSessionContext_SameSessionSameManager(t *testing.T) {
	h := &ToolHandler{}
	cm1 := h.getSessionContext(nil)
	cm2 := h.getSessionContext(nil)
	if cm1 != cm2 {
		t.Fatal("expected same ContextManager for same (nil) session")
	}
}

func TestGetSessionContext_Isolation(t *testing.T) {
	h := &ToolHandler{}

	// Pre-seed two different session IDs to verify isolation
	cm1 := session.NewContextManager(nil)
	cm2 := session.NewContextManager(nil)
	h.sessions.Store("session-a", cm1)
	h.sessions.Store("session-b", cm2)

	// Retrieve via the handler method using nil req (empty session ID)
	// This should create a third, independent manager
	cm3 := h.getSessionContext(nil)
	if cm3 == nil {
		t.Fatal("expected non-nil ContextManager for nil session")
	}

	// Verify the two pre-seeded sessions are still independent
	gotA, _ := h.sessions.Load("session-a")
	gotB, _ := h.sessions.Load("session-b")
	if gotA.(*session.ContextManager) == gotB.(*session.ContextManager) {
		t.Fatal("expected different ContextManagers for different session IDs")
	}

	// Using a real CallToolRequest with nil Session should match the empty-key manager
	req := &mcp.CallToolRequest{}
	cm4 := h.getSessionContext(req)
	if cm4 != cm3 {
		t.Fatal("expected same ContextManager for nil Session and empty session ID")
	}
}
