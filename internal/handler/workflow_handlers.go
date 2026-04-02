package handler

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gitopia/gitopia-mcp-server/internal/constants"
	"github.com/gitopia/gitopia-mcp-server/internal/fs"
	"github.com/gitopia/gitopia-mcp-server/internal/git"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Workflow Handlers ----

type CreateFeatureBranchPRParams struct {
	RepoPath      string          `json:"repo_path" jsonschema:"Repository path relative to workspace (e.g. 'myrepo')"`
	Owner         string          `json:"owner"`
	Name          string          `json:"name"`
	BranchName    string          `json:"branch_name"`
	BaseBranch    string          `json:"base_branch,omitempty"`
	Files         []fs.FileChange `json:"files"`
	CommitMessage string          `json:"commit_message"`
	PRTitle       string          `json:"pr_title"`
	PRDescription string          `json:"pr_description"`
	Assignees     []string        `json:"assignees,omitempty"`
	Labels        []uint64        `json:"labels,omitempty"`
}

func (h *ToolHandler) CreateFeatureBranchPR(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreateFeatureBranchPRParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("create_feature_branch_pr", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}
	walletData, err := w.Export()
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to export wallet: %s", err)
	}
	if err := h.AuthMgr.SetupAuth(walletData, w.Address()); err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to setup git auth: %s", err)
	}
	defer func() { _ = h.AuthMgr.CleanupAuth() }()

	// Resolve repository path
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}

	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Provided path is not a git repository")
	}

	baseBranch := p.BaseBranch
	if baseBranch == "" {
		baseBranch = "main"
	}
	logging.Infof("Creating branch %s from %s", p.BranchName, baseBranch)

	if err := h.GitClient.Checkout(ctx, absRepoPath, baseBranch); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to checkout base branch %s: %s", baseBranch, err)
	}
	if err := h.GitClient.Pull(ctx, absRepoPath, git.WithRemoteToPull("origin"), git.WithBranchToPull(baseBranch)); err != nil {
		logging.Warnf("failed to pull latest changes: %v", err)
	}

	if err := h.GitClient.CreateBranch(ctx, absRepoPath, p.BranchName); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to create new branch %s: %s", p.BranchName, err)
	}
	if err := h.GitClient.Checkout(ctx, absRepoPath, p.BranchName); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to checkout new branch %s: %s", p.BranchName, err)
	}

	for _, file := range p.Files {
		if err := fs.ApplyChange(absRepoPath, file); err != nil {
			return toolErrorf(ErrFileSystem, "Failed to apply changes to %s: %s", file.Path, err)
		}
	}

	if err := h.GitClient.Add(ctx, absRepoPath); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to stage changes: %s", err)
	}

	commitHash, err := h.GitClient.Commit(ctx, absRepoPath, p.CommitMessage)
	if err != nil {
		return toolErrorf(ErrGitOperation, "Failed to commit changes: %s", err)
	}

	pushOptions := []git.PushOption{
		git.WithRemote("origin"),
		git.WithBranchToPush(p.BranchName),
		git.WithSetUpstream(true),
	}
	if err := h.GitClient.Push(ctx, absRepoPath, pushOptions...); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to push branch: %s", err)
	}

	prNumber, err := retryOnSequenceMismatch(func() (int, error) {
		return h.GClient.CreatePullRequest(ctx, w, p.Owner, p.Name, p.PRTitle, p.PRDescription, p.BranchName, baseBranch, p.Assignees, p.Labels, nil)
	})

	if err != nil {
		// Partial success: branch created and pushed, but PR creation failed
		return toolSuccess("Created feature branch (PR creation failed)", map[string]any{
			"branch":   p.BranchName,
			"commit":   commitHash[:8],
			"pr_error": err.Error(),
		})
	}

	url := fmt.Sprintf("https://gitopia.com/%s/%s/pulls/%d", p.Owner, p.Name, prNumber)
	return toolSuccess("Created feature branch and pull request", map[string]any{
		"branch":    p.BranchName,
		"commit":    commitHash[:8],
		"pull_iid": prNumber,
		"url":       url,
	})
}

type UpdateFeatureBranchParams struct {
	RepoPath      string          `json:"repo_path" jsonschema:"Repository path relative to workspace (e.g. 'myrepo')"`
	BranchName    string          `json:"branch_name"`
	Files         []fs.FileChange `json:"files"`
	CommitMessage string          `json:"commit_message"`
}

func (h *ToolHandler) UpdateFeatureBranch(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p UpdateFeatureBranchParams,
) (*mcp.CallToolResult, any, error) {
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}
	walletData, err := w.Export()
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to export wallet: %s", err)
	}
	if err := h.AuthMgr.SetupAuth(walletData, w.Address()); err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to setup git auth: %s", err)
	}
	defer func() { _ = h.AuthMgr.CleanupAuth() }()

	// Resolve repository path
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	if err := h.GitClient.Checkout(ctx, absRepoPath, p.BranchName); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to checkout branch %s: %s", p.BranchName, err)
	}
	filesChanged := 0
	for _, file := range p.Files {
		if err := fs.ApplyChange(absRepoPath, file); err != nil {
			return toolErrorf(ErrFileSystem, "Failed to apply changes to %s: %s", file.Path, err)
		}
		filesChanged++
	}
	if err := h.GitClient.Add(ctx, absRepoPath); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to stage changes: %s", err)
	}
	commitHash, err := h.GitClient.Commit(ctx, absRepoPath, p.CommitMessage)
	if err != nil {
		return toolErrorf(ErrGitOperation, "Failed to commit changes: %s", err)
	}
	pushOptions := []git.PushOption{
		git.WithRemote("origin"),
		git.WithBranchToPush(p.BranchName),
	}
	if err := h.GitClient.Push(ctx, absRepoPath, pushOptions...); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to push branch: %s", err)
	}
	return toolSuccess("Updated feature branch", map[string]any{
		"branch":        p.BranchName,
		"commit":        commitHash[:8],
		"files_changed": filesChanged,
	})
}

// ---- Bootstrap Repository Workflow ----

type BootstrapRepoParams struct {
	OwnerId           string `json:"owner_id"`
	Name              string `json:"name"`
	LocalPath         string `json:"local_path,omitempty"`
	Description       string `json:"description,omitempty"`
	CreateReadme      bool   `json:"create_readme"`
	CreateGitignore   bool   `json:"create_gitignore"`
	InitialBranch     string `json:"initial_branch,omitempty"`
	GitignoreTemplate string `json:"gitignore_template,omitempty"`
}

func (h *ToolHandler) BootstrapRepo(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p BootstrapRepoParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("bootstrap_repo", p)
	}

	// Set defaults and normalize path
	localPath := p.LocalPath
	if localPath == "" {
		localPath = p.Name
	} else {
		// If the client provided a full /tmp path, extract just the directory name
		// This handles cases where MCP clients auto-generate temp paths
		if strings.HasPrefix(localPath, "/tmp/") {
			localPath = filepath.Base(localPath)
		}
	}

	branch := p.InitialBranch
	if branch == "" {
		branch = "main"
	}

	// Get wallet for authentication
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}

	// Setup git auth
	walletData, err := w.Export()
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to export wallet: %s", err)
	}
	if err := h.AuthMgr.SetupAuth(walletData, w.Address()); err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to setup git auth: %s", err)
	}
	defer func() { _ = h.AuthMgr.CleanupAuth() }()

	// Step 1: Create remote repository
	logging.Infof("Creating remote repository %s/%s", p.OwnerId, p.Name)
	resp, err := h.GClient.CreateRepository(ctx, w, p.OwnerId, p.Name, p.Description)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to create remote repository: %s", err)
	}
	repoURL := fmt.Sprintf("https://gitopia.com/%s/%s", resp.RepositoryId.Id, resp.RepositoryId.Name)

	// Step 2: Initialize local repository
	absLocalPath, err := h.ResolvePath(localPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve local path '%s': %s", localPath, err)
	}

	logging.Infof("Initializing local repository at %s", absLocalPath)
	remoteURL := fmt.Sprintf("gitopia://%s/%s", p.OwnerId, p.Name)
	if err := h.GitClient.Init(ctx, absLocalPath, remoteURL); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to initialize local repository: %s", err)
	}

	// Step 3: Create initial files
	if p.CreateReadme {
		readmeContent := fmt.Sprintf("# %s\n\n%s\n", p.Name, p.Description)
		if p.Description == "" {
			readmeContent = fmt.Sprintf("# %s\n\nA new repository created with Gitopia MCP Server.\n", p.Name)
		}
		if err := fs.Write(absLocalPath, "README.md", readmeContent); err != nil {
			return toolErrorf(ErrFileSystem, "Failed to create README.md: %s", err)
		}
	}

	if p.CreateGitignore {
		tpl := constants.GetGitignoreTemplate(p.GitignoreTemplate)
		if err := fs.Write(absLocalPath, ".gitignore", tpl); err != nil {
			return toolErrorf(ErrFileSystem, "Failed to create .gitignore: %s", err)
		}
	}

	// Step 4: Create and checkout initial branch
	logging.Infof("Creating and checking out initial branch %s", branch)
	if err := h.GitClient.Checkout(ctx, absLocalPath, branch, git.WithCreateBranch(true)); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to create and checkout branch %s: %s", branch, err)
	}

	if err := h.GitClient.Add(ctx, absLocalPath); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to stage initial files: %s", err)
	}

	commitMessage := "Initial commit"
	commitHash, err := h.GitClient.Commit(ctx, absLocalPath, commitMessage)
	if err != nil {
		return toolErrorf(ErrGitOperation, "Failed to create initial commit: %s", err)
	}

	// Step 5: Push to remote with upstream
	logging.Infof("Pushing initial commit to remote")
	pushOptions := []git.PushOption{
		git.WithRemote("origin"),
		git.WithBranchToPush(branch),
		git.WithSetUpstream(true),
	}
	if err := h.GitClient.Push(ctx, absLocalPath, pushOptions...); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to push initial commit: %s", err)
	}

	return toolSuccess("Bootstrapped repository", map[string]any{
		"url":        repoURL,
		"local_path": localPath,
		"branch":     branch,
		"commit":     commitHash[:8],
	})
}
