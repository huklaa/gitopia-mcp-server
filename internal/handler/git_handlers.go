package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/gitopia/gitopia-mcp-server/internal/git"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/logging"
	"github.com/gitopia/gitopia-mcp-server/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Git Operation Handlers ----

type GitCloneParams struct {
	RepoURL   string `json:"repo_url" jsonschema:"gitopia:// URL or owner/name format"`
	LocalPath string `json:"local_path" jsonschema:"Path relative to workspace root"`
	Branch    string `json:"branch,omitempty" jsonschema:"Specific branch to clone"`
	Depth     int    `json:"depth,omitempty" jsonschema:"Shallow clone depth"`
}

func (h *ToolHandler) GitClone(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p GitCloneParams,
) (*mcp.CallToolResult, any, error) {
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}
	walletData, err := w.Export()
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to export wallet: %s", err)
	}
	if err := h.AuthMgr.SetupAuth(walletData, w.Address()); err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to setup git auth: %s", err)
	}
	defer func() { _ = h.AuthMgr.CleanupAuth() }()

	normalizedURL := git.NormalizeGitopiaURL(p.RepoURL)
	if !git.ValidateGitopiaURL(normalizedURL) {
		return toolErrorf(ErrValidation, "Invalid Gitopia URL: %s. Use gitopia://owner/repo format.", p.RepoURL)
	}
	absolutePath, err := h.ResolvePath(p.LocalPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve path '%s': %s", p.LocalPath, err)
	}
	if _, err := os.Stat(absolutePath); err == nil {
		return toolErrorf(ErrConflict, "Target directory already exists: %s", p.LocalPath)
	}
	var options []git.CloneOption
	if p.Branch != "" {
		options = append(options, git.WithBranch(p.Branch))
	}
	if p.Depth > 0 {
		options = append(options, git.WithDepth(p.Depth))
	}
	if err := h.GitClient.Clone(ctx, normalizedURL, absolutePath, options...); err != nil {
		if gitErr, ok := err.(*git.GitError); ok {
			return toolErrorf(ErrGitOperation, "Clone failed: %s", gitErr.UserFriendlyMessage())
		}
		return toolErrorf(ErrGitOperation, "Clone failed: %s", err)
	}
	logging.Infof("Successfully cloned %s to %s", normalizedURL, p.LocalPath)
	return toolSuccess("Cloned repository", map[string]any{
		"url":        normalizedURL,
		"local_path": p.LocalPath,
		"branch":     p.Branch,
	})
}

type GitStatusParams struct {
	RepoPath string `json:"repo_path" jsonschema:"Repository path relative to workspace"`
}

func (h *ToolHandler) GitStatus(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p GitStatusParams,
) (*mcp.CallToolResult, any, error) {
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Not a git repository: %s", p.RepoPath)
	}
	status, err := h.GitClient.Status(ctx, absRepoPath)
	if err != nil {
		if gitErr, ok := err.(*git.GitError); ok {
			return toolErrorf(ErrGitOperation, "Status failed: %s", gitErr.UserFriendlyMessage())
		}
		return toolErrorf(ErrGitOperation, "Status failed: %s", err)
	}
	return toolSuccessData("Git status", status)
}

type GitPushParams struct {
	RepoPath    string `json:"repo_path" jsonschema:"Repository path relative to workspace"`
	Remote      string `json:"remote,omitempty" jsonschema:"Remote name (default: origin)"`
	Branch      string `json:"branch,omitempty" jsonschema:"Branch to push"`
	Force       bool   `json:"force,omitempty" jsonschema:"Force push"`
	SetUpstream bool   `json:"set_upstream,omitempty" jsonschema:"Set upstream tracking"`
	DryRun      bool   `json:"dry_run,omitempty" jsonschema:"Dry run"`
}

func (h *ToolHandler) GitPush(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p GitPushParams,
) (*mcp.CallToolResult, any, error) {
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}
	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Not a git repository: %s", p.RepoPath)
	}
	walletData, err := w.Export()
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to export wallet: %s", err)
	}
	if err := h.AuthMgr.SetupAuth(walletData, w.Address()); err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to setup git auth: %s", err)
	}
	defer func() { _ = h.AuthMgr.CleanupAuth() }()

	branchToPush := p.Branch
	if branchToPush == "" {
		currentBranch, err := h.GitClient.GetCurrentBranch(ctx, absRepoPath)
		if err != nil {
			return toolErrorf(ErrGitOperation, "Could not determine current branch: %s", err)
		}
		branchToPush = currentBranch
	}
	var options []git.PushOption
	remote := "origin"
	if p.Remote != "" {
		remote = p.Remote
	}
	options = append(options, git.WithRemote(remote), git.WithBranchToPush(branchToPush))
	if p.Force {
		options = append(options, git.WithForce(true))
	}
	if p.SetUpstream {
		options = append(options, git.WithSetUpstream(true))
	}
	if p.DryRun {
		options = append(options, git.WithDryRun(true))
	}
	if err := h.GitClient.Push(ctx, absRepoPath, options...); err != nil {
		if gitErr, ok := err.(*git.GitError); ok {
			return toolErrorf(ErrGitOperation, "Push failed: %s", gitErr.UserFriendlyMessage())
		}
		return toolErrorf(ErrGitOperation, "Push failed: %s", err)
	}
	message := "Pushed to remote"
	if p.DryRun {
		message = "Dry run completed - no changes were pushed"
	}
	logging.Infof("Push completed for %s", p.RepoPath)
	return toolSuccess(message, map[string]any{"remote": remote, "branch": branchToPush})
}

type CommitAndPushChangesParams struct {
	RepoPath      string   `json:"repo_path" jsonschema:"Repository path relative to workspace"`
	CommitMessage string   `json:"commit_message" jsonschema:"Commit message"`
	Files         []string `json:"files,omitempty" jsonschema:"Specific files to stage (empty for all)"`
	Branch        string   `json:"branch,omitempty" jsonschema:"Target branch (current if empty)"`
	CreateBranch  bool     `json:"create_branch,omitempty" jsonschema:"Create branch if it does not exist"`
}

func (h *ToolHandler) CommitAndPushChanges(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CommitAndPushChangesParams,
) (*mcp.CallToolResult, any, error) {
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}
	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Not a git repository: %s", p.RepoPath)
	}
	walletData, err := w.Export()
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to export wallet: %s", err)
	}
	if err := h.AuthMgr.SetupAuth(walletData, w.Address()); err != nil {
		return toolErrorf(ErrAuthFailed, "Failed to setup git auth: %s", err)
	}
	defer func() { _ = h.AuthMgr.CleanupAuth() }()

	status, err := h.GitClient.Status(ctx, absRepoPath)
	if err != nil {
		return toolErrorf(ErrGitOperation, "Failed to get status: %s", err)
	}
	if status.IsClean() {
		return toolSuccess("No changes to commit. Working tree is clean.", nil)
	}
	if p.CreateBranch && p.Branch != "" && p.Branch != status.Branch {
		if err := h.GitClient.CreateBranch(ctx, absRepoPath, p.Branch); err != nil {
			return toolErrorf(ErrGitOperation, "Failed to create branch: %s", err)
		}
	}
	if len(p.Files) > 0 {
		if err := h.GitClient.Add(ctx, absRepoPath, p.Files...); err != nil {
			return toolErrorf(ErrGitOperation, "Failed to stage files: %s", err)
		}
	} else {
		if err := h.GitClient.Add(ctx, absRepoPath); err != nil {
			return toolErrorf(ErrGitOperation, "Failed to stage changes: %s", err)
		}
	}
	commitHash, err := h.GitClient.Commit(ctx, absRepoPath, p.CommitMessage)
	if err != nil {
		return toolErrorf(ErrGitOperation, "Failed to commit: %s", err)
	}
	var pushOptions []git.PushOption
	pushOptions = append(pushOptions, git.WithRemote("origin"))
	branchToPush := p.Branch
	if branchToPush == "" {
		branchToPush = status.Branch
	}
	pushOptions = append(pushOptions, git.WithBranchToPush(branchToPush), git.WithSetUpstream(true))
	if err := h.GitClient.Push(ctx, absRepoPath, pushOptions...); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to push: %s", err)
	}
	logging.Infof("Committed and pushed changes to %s", p.RepoPath)
	filesCount := len(status.Staged) + len(status.Modified) + len(status.Untracked)
	return toolSuccess("Committed and pushed changes", map[string]any{
		"commit": commitHash[:8],
		"branch": status.Branch,
		"files":  filesCount,
	})
}

type GitInitParams struct {
	Path      string `json:"path" jsonschema:"Repository path relative to workspace root"`
	RemoteURL string `json:"remote_url,omitempty" jsonschema:"The gitopia URL to set as origin remote"`
}

func (h *ToolHandler) GitInit(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p GitInitParams,
) (*mcp.CallToolResult, any, error) {
	repoPath, err := h.ResolvePath(p.Path)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve path '%s': %s", p.Path, err)
	}
	if _, err := os.Stat(repoPath); err == nil {
		return toolErrorf(ErrConflict, "Target path already exists: %s", repoPath)
	}
	normalizedURL := ""
	if p.RemoteURL != "" {
		normalizedURL = git.NormalizeGitopiaURL(p.RemoteURL)
		if !git.ValidateGitopiaURL(normalizedURL) {
			return toolErrorf(ErrValidation, "Invalid Gitopia URL: %s", p.RemoteURL)
		}
	}
	if err := h.GitClient.Init(ctx, repoPath, normalizedURL); err != nil {
		return toolErrorf(ErrGitOperation, "Failed to initialize repository: %s", err)
	}
	if err := h.GitClient.Checkout(ctx, repoPath, "main", git.WithCreateBranch(true)); err != nil {
		logging.Warnf("Could not create main branch (using default): %v", err)
	} else {
		logging.Infof("Set default branch to 'main'")
	}
	logging.Infof("Successfully initialized repository at %s", repoPath)
	return toolSuccess(fmt.Sprintf("Initialized git repository at '%s'", p.Path), nil)
}

type SearchCodeParams struct {
	RepoPath      string `json:"repo_path" jsonschema:"Repository path relative to workspace"`
	Query         string `json:"query" jsonschema:"Search query string or regex pattern"`
	CaseSensitive bool   `json:"case_sensitive,omitempty" jsonschema:"Whether search should be case-sensitive"`
}

func (h *ToolHandler) SearchCode(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p SearchCodeParams,
) (*mcp.CallToolResult, any, error) {
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Not a git repository: %s", p.RepoPath)
	}
	matches, err := search.CodeSearch(absRepoPath, p.Query, p.CaseSensitive)
	if err != nil {
		return toolErrorf(ErrGitOperation, "Code search failed: %s", err)
	}
	if len(matches) == 0 {
		return toolSuccessData("No matches found", []any{})
	}
	return toolSuccessData(fmt.Sprintf("Found %d matches", len(matches)), matches)
}

type GitAddParams struct {
	RepoPath string   `json:"repo_path" jsonschema:"Repository path relative to workspace"`
	Files    []string `json:"files" jsonschema:"Files to stage. Use '.' to stage all changes."`
}

func (h *ToolHandler) GitAdd(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p GitAddParams,
) (*mcp.CallToolResult, any, error) {
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Not a git repository: %s", p.RepoPath)
	}
	if err := h.GitClient.Add(ctx, absRepoPath, p.Files...); err != nil {
		if gitErr, ok := err.(*git.GitError); ok {
			return toolErrorf(ErrGitOperation, "git add failed: %s", gitErr.UserFriendlyMessage())
		}
		return toolErrorf(ErrGitOperation, "git add failed: %s", err)
	}
	logging.Infof("Added %d files to the index in %s", len(p.Files), p.RepoPath)
	return toolSuccess(fmt.Sprintf("Staged %d files", len(p.Files)), map[string]any{"files": p.Files})
}

type GitCommitParams struct {
	RepoPath    string `json:"repo_path" jsonschema:"Repository path relative to workspace"`
	Message     string `json:"message" jsonschema:"The commit message"`
	AuthorName  string `json:"author_name,omitempty" jsonschema:"Optional author name"`
	AuthorEmail string `json:"author_email,omitempty" jsonschema:"Optional author email"`
}

func (h *ToolHandler) GitCommit(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p GitCommitParams,
) (*mcp.CallToolResult, any, error) {
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Not a git repository: %s", p.RepoPath)
	}
	var commitOptions []git.CommitOption
	if p.AuthorName != "" && p.AuthorEmail != "" {
		author := fmt.Sprintf("%s <%s>", p.AuthorName, p.AuthorEmail)
		commitOptions = append(commitOptions, git.WithAuthor(author))
	}
	commitHash, err := h.GitClient.Commit(ctx, absRepoPath, p.Message, commitOptions...)
	if err != nil {
		if gitErr, ok := err.(*git.GitError); ok {
			return toolErrorf(ErrGitOperation, "git commit failed: %s", gitErr.UserFriendlyMessage())
		}
		return toolErrorf(ErrGitOperation, "git commit failed: %s", err)
	}
	logging.Infof("Created commit in %s: %s", p.RepoPath, commitHash)
	return toolSuccess("Created commit", map[string]any{"commit": commitHash})
}

type CreateFeatureBranchParams struct {
	RepoPath   string `json:"repo_path"`
	BranchName string `json:"branch_name"`
	BaseBranch string `json:"base_branch,omitempty"`
}

func (h *ToolHandler) CreateFeatureBranch(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreateFeatureBranchParams,
) (*mcp.CallToolResult, any, error) {
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Not a git repository: %s", p.RepoPath)
	}
	var opts []git.BranchOption
	if p.BaseBranch != "" {
		opts = append(opts, git.WithStartPoint(p.BaseBranch))
	}
	if err := h.GitClient.CreateBranch(ctx, absRepoPath, p.BranchName, opts...); err != nil {
		if gitErr, ok := err.(*git.GitError); ok {
			return toolErrorf(ErrGitOperation, "Failed to create branch: %s", gitErr.UserFriendlyMessage())
		}
		return toolErrorf(ErrGitOperation, "Failed to create branch: %s", err)
	}
	logging.Infof("Created feature branch '%s' in %s", p.BranchName, p.RepoPath)
	return toolSuccess(fmt.Sprintf("Created branch '%s'", p.BranchName), map[string]any{"branch": p.BranchName})
}

type SyncWithRemoteParams struct {
	RepoPath string `json:"repo_path"`
	Branch   string `json:"branch,omitempty"`
	Remote   string `json:"remote,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}

func (h *ToolHandler) SyncWithRemote(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p SyncWithRemoteParams,
) (*mcp.CallToolResult, any, error) {
	absRepoPath, err := h.ResolvePath(p.RepoPath)
	if err != nil {
		return toolErrorf(ErrPathTraversal, "Failed to resolve repo path '%s': %s", p.RepoPath, err)
	}
	if !h.GitClient.IsRepository(absRepoPath) {
		return toolErrorf(ErrNotRepository, "Not a git repository: %s", p.RepoPath)
	}
	var pullOptions []git.PullOption
	if p.Remote != "" {
		pullOptions = append(pullOptions, git.WithRemoteToPull(p.Remote))
	}
	if p.Branch != "" {
		pullOptions = append(pullOptions, git.WithBranchToPull(p.Branch))
	}
	if p.Strategy == "rebase" {
		pullOptions = append(pullOptions, git.WithRebase(true))
	}
	if err := h.GitClient.Pull(ctx, absRepoPath, pullOptions...); err != nil {
		if gitErr, ok := err.(*git.GitError); ok {
			return toolErrorf(ErrGitOperation, "Failed to sync with remote: %s", gitErr.UserFriendlyMessage())
		}
		return toolErrorf(ErrGitOperation, "Failed to sync with remote: %s", err)
	}
	logging.Infof("Successfully synced repository %s with remote", p.RepoPath)
	return toolSuccess("Synced with remote", nil)
}
