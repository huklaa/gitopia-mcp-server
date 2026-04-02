package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/config"
	"github.com/gitopia/gitopia-mcp-server/internal/git"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/handler"
	"github.com/gitopia/gitopia-mcp-server/internal/logging"
	"github.com/gitopia/gitopia-mcp-server/internal/workspace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	// Handle --version early, before any initialization.
	// Dockerfile HEALTHCHECK depends on this flag working.
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Println(version)
			os.Exit(0)
		}
	}

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "gitopia-mcp-server",
		Version: version,
	}, nil)

	// Load configuration
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		logging.Fatalf("failed to load configuration: %v", err)
	}

	// Configure logger
	logging.ConfigureLogger(logging.Config{Level: cfg.Logging.Level})

	// Initialize workspace manager
	workspaceMgr, err := workspace.NewManager(cfg.Workspace)
	if err != nil {
		logging.Fatalf("failed to create workspace manager: %v", err)
	}

	// Initialize shared clients and managers
	var grpcEndpoints []string
	if len(cfg.GRPCEndpoints) > 0 {
		grpcEndpoints = cfg.GRPCEndpoints
	} else if ep := os.Getenv("GITOPIA_GRPC_ENDPOINT"); ep != "" {
		grpcEndpoints = []string{ep}
	} else {
		grpcEndpoints = []string{gitopia.DefaultGRPC}
	}
	gClient, err := gitopia.New(context.Background(), grpcEndpoints...)
	if err != nil {
		logging.Fatalf("failed to create gitopia client: %v", err)
	}
	defer func() { _ = gClient.Close() }()

	gitClient := git.NewClient()
	authMgr := git.NewAuthManager()

	// Setup git configuration from environment and config
	gitName := os.Getenv("GIT_USER_NAME")
	gitEmail := os.Getenv("GIT_USER_EMAIL")
	if gitName == "" {
		gitName = cfg.Git.DefaultUserName
	}
	if gitEmail == "" {
		gitEmail = cfg.Git.DefaultUserEmail
	}
	if err := git.SetupGitConfig(context.Background(), gitName, gitEmail); err != nil {
		logging.Warnf("Failed to setup git config: %v", err)
	}

	// Create a new ToolHandler with the initialized clients and workspace manager
	h := handler.New(gClient, gitClient, authMgr, workspaceMgr)
	h.TrustLvl = handler.ParseTrustLevel(cfg.TrustLevel)
	h.ChainLimiter = handler.NewRateLimiter(cfg.RateLimit.ChainTxPerMinute, cfg.RateLimit.ChainTxPerHour)
	h.DryRunMode = cfg.DryRun
	h.ActiveToolsets = handler.ParseToolsets(cfg.Toolsets)
	h.ApprovalMode = cfg.ApprovalMode
	if h.ApprovalMode {
		ttl := 5 * time.Minute
		if cfg.ApprovalTTL != "" {
			if parsed, err := time.ParseDuration(cfg.ApprovalTTL); err == nil {
				ttl = parsed
			}
		}
		h.PendingTxs = handler.NewPendingTxStore(ttl)
	}

	// Register all tools, prompts, and resource templates
	handler.RegisterAll(srv, h)

	// Select transport based on configuration
	switch cfg.Transport {
	case "http":
		port := cfg.HTTPPort
		if port == "" {
			port = "8080"
		}
		httpHandler := mcp.NewStreamableHTTPHandler(
			func(r *http.Request) *mcp.Server { return srv },
			&mcp.StreamableHTTPOptions{},
		)
		logging.Infof("Starting HTTP transport on :%s", port)
		if err := (&http.Server{Addr: ":" + port, Handler: httpHandler}).ListenAndServe(); err != nil {
			logging.Fatal(err)
		}
	default: // "stdio"
		if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			logging.Fatal(err)
		}
	}
}
