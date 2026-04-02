package handler

import (
	"strings"
	"sync"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/git"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/logging"
	"github.com/gitopia/gitopia-mcp-server/internal/session"
	"github.com/gitopia/gitopia-mcp-server/internal/workspace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolHandler holds shared dependencies for all tool handlers.
// By making handlers methods on this struct, we can easily share
// state and dependencies like clients and managers without using globals.
type ToolHandler struct {
	GClient      *gitopia.Client
	GitClient    *git.Client
	AuthMgr      *git.AuthManager
	WorkspaceMgr *workspace.Manager
	TrustLvl     TrustLevel
	ChainLimiter *RateLimiter
	DryRunMode   bool
	ApprovalMode    bool
	PendingTxs      *PendingTxStore
	ActiveToolsets   map[Toolset]bool
	sessions     sync.Map // map[string]*session.ContextManager — keyed by session ID
}

// New creates a new ToolHandler.
func New(g *gitopia.Client, git *git.Client, auth *git.AuthManager, ws *workspace.Manager) *ToolHandler {
	return &ToolHandler{
		GClient:      g,
		GitClient:    git,
		AuthMgr:      auth,
		WorkspaceMgr: ws,
	}
}

// getSessionContext returns the ContextManager for the given request's session.
// If no session is present (e.g. stdio transport), all requests share a single
// ContextManager keyed by the empty string.
func (h *ToolHandler) getSessionContext(req *mcp.CallToolRequest) *session.ContextManager {
	var sid string
	if req != nil && req.Session != nil {
		sid = req.Session.ID()
	}
	if v, ok := h.sessions.Load(sid); ok {
		return v.(*session.ContextManager)
	}
	cm := session.NewContextManager(h.GClient)
	actual, _ := h.sessions.LoadOrStore(sid, cm)
	return actual.(*session.ContextManager)
}

// ResolvePath resolves a relative path to an absolute workspace path
func (h *ToolHandler) ResolvePath(relativePath string) (string, error) {
	return h.WorkspaceMgr.ResolvePath(relativePath)
}

// retryOnSequenceMismatch retries fn up to 3 times when the error contains
// "account sequence mismatch", using linear back-off.
func retryOnSequenceMismatch(fn func() (int, error)) (int, error) {
	const maxRetries = 3
	const retryDelay = 2 * time.Second
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		if strings.Contains(err.Error(), "account sequence mismatch") {
			logging.Warnf("Account sequence mismatch (attempt %d/%d). Retrying...", i+1, maxRetries)
			time.Sleep(retryDelay * time.Duration(i+1))
			lastErr = err
			continue
		}
		return 0, err
	}
	return 0, lastErr
}
