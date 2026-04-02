//go:build integration

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/config"
	"github.com/gitopia/gitopia-mcp-server/internal/git"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/handler"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	"github.com/gitopia/gitopia-mcp-server/internal/workspace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// testState holds shared mutable state across all test tiers.
// Each tier populates fields that downstream tiers consume.
type testState struct {
	session *mcp.ClientSession
	ctx     context.Context
	runID   string // "e2e-{unix_timestamp}" for unique names
	gClient *gitopia.Client // direct gRPC client for chain queries in tests

	// Populated by tier 0
	username string
	address  string

	// Populated by tier 2
	repoName   string
	repoOwner  string
	repoID     uint64 // numeric repo ID from chain
	repoPushed bool   // whether the initial push to remote succeeded

	// Populated by tier 4
	issueIID uint64

	// Populated by tier 5
	prIID uint64

	// Populated by tier 6
	bountyID uint64

	// Populated by tier 7
	daoName            string
	groupID            uint64
	groupPolicyAddress string
	proposalID         uint64

	// Populated by tier B
	labelID uint64

	// Populated by tier 8 (approval mode)
	approvalSession *mcp.ClientSession
	approvalHandler *handler.ToolHandler
}

var state *testState

func TestMain(m *testing.M) {
	// Auto-generate a mnemonic if one isn't provided.
	// This gives each test run a fresh wallet identity.
	if os.Getenv("GITOPIA_MNEMONIC") == "" {
		mnemonic, err := signing.GenerateMnemonic()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate mnemonic: %v\n", err)
			os.Exit(1)
		}
		os.Setenv("GITOPIA_MNEMONIC", mnemonic)
		fmt.Fprintln(os.Stderr, "Auto-generated a fresh wallet mnemonic for E2E tests")
	}

	ctx := context.Background()
	state = &testState{
		ctx:   ctx,
		runID: fmt.Sprintf("e2e-%d", time.Now().Unix()),
	}

	// Create real gRPC client
	var endpoints []string
	if ep := os.Getenv("GITOPIA_GRPC_ENDPOINTS"); ep != "" {
		endpoints = []string{ep}
	} else {
		endpoints = []string{gitopia.DefaultGRPC}
	}
	gClient, err := gitopia.New(ctx, endpoints...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create gitopia client: %v\n", err)
		os.Exit(1)
	}
	defer gClient.Close()

	// Create workspace in temp directory
	tmpDir, err := os.MkdirTemp("", "e2e-workspace-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	wsMgr, err := workspace.NewManager(&config.WorkspaceConfig{BasePath: tmpDir})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create workspace manager: %v\n", err)
		os.Exit(1)
	}

	// Create git client and auth manager
	gitClient := git.NewClient()
	authMgr := git.NewAuthManager()

	// Setup git config for E2E identity
	if err := git.SetupGitConfig(ctx, "E2E Test", "e2e@gitopia.test"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to setup git config: %v\n", err)
	}

	state.gClient = gClient

	// Create tool handler with generous rate limits
	h := handler.New(gClient, gitClient, authMgr, wsMgr)
	h.TrustLvl = handler.TrustChainWrite
	h.ChainLimiter = handler.NewRateLimiter(120, 3600) // generous for testing

	// Create MCP server and register all tools
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "gitopia-e2e-test",
		Version: "test",
	}, nil)
	handler.RegisterAll(srv, h)

	// Wire up in-memory transport
	t1, t2 := mcp.NewInMemoryTransports()
	serverSession, err := srv.Connect(ctx, t1, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect server: %v\n", err)
		os.Exit(1)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "gitopia-e2e-client",
		Version: "test",
	}, nil)
	clientSession, err := client.Connect(ctx, t2, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect client: %v\n", err)
		os.Exit(1)
	}
	defer clientSession.Close()
	state.session = clientSession

	// Setup approval mode server for tier 8
	setupApprovalServer(ctx, gClient, gitClient, authMgr, wsMgr)

	code := m.Run()
	os.Exit(code)
}

func setupApprovalServer(ctx context.Context, gClient *gitopia.Client, gitClient *git.Client, authMgr *git.AuthManager, wsMgr *workspace.Manager) {
	h := handler.New(gClient, gitClient, authMgr, wsMgr)
	h.TrustLvl = handler.TrustChainWrite
	h.ChainLimiter = handler.NewRateLimiter(120, 3600)
	h.ApprovalMode = true
	h.PendingTxs = handler.NewPendingTxStore(5 * time.Minute)
	state.approvalHandler = h

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "gitopia-e2e-approval",
		Version: "test",
	}, nil)
	handler.RegisterAll(srv, h)

	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, t1, nil); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to connect approval server: %v\n", err)
		return
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "gitopia-e2e-approval-client",
		Version: "test",
	}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to connect approval client: %v\n", err)
		return
	}
	state.approvalSession = session
}
