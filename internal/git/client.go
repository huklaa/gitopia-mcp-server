package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/logging"
)

// Client wraps git CLI operations
type Client struct {
	// GitPath is the path to the git executable (default: "git")
	GitPath string
	// Timeout for git operations (default: 30 seconds)
	Timeout time.Duration
}

// NewClient creates a new git client
func NewClient() *Client {
	return &Client{
		GitPath: "git",
		Timeout: 30 * time.Second,
	}
}

// GitError represents a git command error with detailed information
type GitError struct {
	Command  string
	Args     []string
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git %s failed (exit %d): %s",
		strings.Join(e.Args, " "), e.ExitCode, e.Stderr)
}

// UserFriendlyMessage returns a user-friendly error message
func (e *GitError) UserFriendlyMessage() string {
	stderr := strings.ToLower(e.Stderr)
	switch {
	case strings.Contains(stderr, "authentication failed"):
		return "Authentication failed. Please check your Gitopia wallet configuration."
	case strings.Contains(stderr, "permission denied"):
		return "Permission denied. You may not have access to this repository."
	case strings.Contains(stderr, "repository not found"):
		return "Repository not found. Please check the repository URL."
	case strings.Contains(stderr, "connection refused") || strings.Contains(stderr, "network"):
		return "Network error. Please check your internet connection."
	case strings.Contains(stderr, "not a git repository"):
		return "The specified path is not a git repository."
	case strings.Contains(stderr, "working tree clean"):
		return "No changes to commit. Working tree is clean."
	case strings.Contains(stderr, "conflict"):
		return "Merge conflict detected. Please resolve conflicts manually."
	default:
		if e.Stderr != "" {
			return fmt.Sprintf("Git operation failed: %s", e.Stderr)
		}
		return fmt.Sprintf("Git operation failed: %v", e.Err)
	}
}

// runGitCommand executes a git command and returns the result
func (c *Client) runGitCommand(ctx context.Context, workDir string, args ...string) (string, error) {
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, c.GitPath, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	// Ensure the command inherits the current process's environment,
	// which is crucial for git-remote-gitopia to find the wallet path.
	cmd.Env = os.Environ()

	// Log the environment for debugging purposes.
	for _, e := range cmd.Env {
		if strings.HasPrefix(e, "GITOPIA_") {
			logging.Debugf("Git Env: %s", e)
		}
	}

	stdout, err := cmd.Output()
	if err != nil {
		var exitCode int
		var stderr string

		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
			stderr = string(exitError.Stderr)
		}

		// "Everything up-to-date" is not an error — git push returns non-zero
		// exit code when there's nothing to push, but the operation succeeded.
		if strings.Contains(stderr, "Everything up-to-date") {
			return "Everything up-to-date", nil
		}

		return "", &GitError{
			Command:  c.GitPath,
			Args:     args,
			ExitCode: exitCode,
			Stdout:   string(stdout),
			Stderr:   stderr,
			Err:      err,
		}
	}

	return strings.TrimSpace(string(stdout)), nil
}

// IsRepository checks if the given path is a git repository
func (c *Client) IsRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && (info.IsDir() || info.Mode().IsRegular()) // .git can be a file in worktrees
}

// Clone clones a repository
func (c *Client) Clone(ctx context.Context, url, path string, options ...CloneOption) error {
	args := []string{"clone"}

	// Apply options
	opts := &CloneOptions{}
	for _, option := range options {
		option(opts)
	}

	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}
	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}
	if opts.Quiet {
		args = append(args, "--quiet")
	}

	args = append(args, url, path)

	_, err := c.runGitCommand(ctx, "", args...)
	return err
}

// Init initializes a new git repository in the given path.
// If remoteURL is provided, it adds it as the 'origin' remote.
func (c *Client) Init(ctx context.Context, path, remoteURL string) error {
	// 1. Initialize the repository
	if _, err := c.runGitCommand(ctx, "", "init", path); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// 2. Add remote if provided
	if remoteURL != "" {
		if _, err := c.runGitCommand(ctx, path, "remote", "add", "origin", remoteURL); err != nil {
			// Best effort: cleanup the created repo if adding remote fails
			_ = os.RemoveAll(path)
			return fmt.Errorf("git remote add failed: %w", err)
		}
	}

	return nil
}

// Status returns the repository status
func (c *Client) Status(ctx context.Context, repoPath string) (*Status, error) {
	// Get porcelain status for parsing
	output, err := c.runGitCommand(ctx, repoPath, "status", "--porcelain", "--branch")
	if err != nil {
		return nil, err
	}

	status := &Status{
		Staged:    make([]string, 0),
		Modified:  make([]string, 0),
		Untracked: make([]string, 0),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "##") {
			// Branch information
			branchInfo := strings.TrimPrefix(line, "## ")
			if strings.Contains(branchInfo, "...") {
				parts := strings.Split(branchInfo, "...")
				status.Branch = parts[0]
				// TODO: parse ahead/behind counts from tracking info
				_ = parts
			} else {
				status.Branch = branchInfo
			}
			continue
		}

		if len(line) < 3 {
			continue
		}

		staged := line[0]
		modified := line[1]
		filename := line[3:]

		switch {
		case staged != ' ' && staged != '?':
			status.Staged = append(status.Staged, filename)
		case modified != ' ':
			status.Modified = append(status.Modified, filename)
		case staged == '?' && modified == '?':
			status.Untracked = append(status.Untracked, filename)
		}
	}

	return status, nil
}

// Add stages files for commit
func (c *Client) Add(ctx context.Context, repoPath string, files ...string) error {
	if len(files) == 0 {
		files = []string{"."}
	}

	args := append([]string{"add"}, files...)
	_, err := c.runGitCommand(ctx, repoPath, args...)
	return err
}

// Commit creates a new commit
func (c *Client) Commit(ctx context.Context, repoPath, message string, options ...CommitOption) (string, error) {
	args := []string{"commit", "-m", message}

	// Apply options
	opts := &CommitOptions{}
	for _, option := range options {
		option(opts)
	}

	if opts.Author != "" {
		args = append(args, "--author", opts.Author)
	}
	if opts.AllowEmpty {
		args = append(args, "--allow-empty")
	}
	if opts.Amend {
		args = append(args, "--amend")
	}

	_, err := c.runGitCommand(ctx, repoPath, args...)
	if err != nil {
		return "", err
	}

	// Get the commit hash
	hash, err := c.runGitCommand(ctx, repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return hash, nil
}

// Push pushes commits to remote
func (c *Client) Push(ctx context.Context, repoPath string, options ...PushOption) error {
	args := []string{"push"}

	// Apply options
	opts := &PushOptions{}
	for _, option := range options {
		option(opts)
	}

	if opts.Remote != "" {
		args = append(args, opts.Remote)
	}
	if opts.Branch != "" {
		args = append(args, opts.Branch)
	}
	if opts.SetUpstream {
		args = append(args, "--set-upstream")
	}
	if opts.Force {
		args = append(args, "--force")
	}
	if opts.ForceWithLease {
		args = append(args, "--force-with-lease")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	if opts.Tags {
		args = append(args, "--tags")
	}

	_, err := c.runGitCommand(ctx, repoPath, args...)
	return err
}

// Pull fetches and merges from remote
func (c *Client) Pull(ctx context.Context, repoPath string, options ...PullOption) error {
	args := []string{"pull"}

	// Apply options
	opts := &PullOptions{}
	for _, option := range options {
		option(opts)
	}

	if opts.Remote != "" {
		args = append(args, opts.Remote)
	}
	if opts.Branch != "" {
		args = append(args, opts.Branch)
	}
	if opts.Rebase {
		args = append(args, "--rebase")
	}
	if opts.FFOnly {
		args = append(args, "--ff-only")
	}

	_, err := c.runGitCommand(ctx, repoPath, args...)
	return err
}

// CreateBranch creates a new branch
func (c *Client) CreateBranch(ctx context.Context, repoPath, branchName string, options ...BranchOption) error {
	args := []string{"checkout", "-b", branchName}

	// Apply options
	opts := &BranchOptions{}
	for _, option := range options {
		option(opts)
	}

	if opts.StartPoint != "" {
		args = append(args, opts.StartPoint)
	}

	_, err := c.runGitCommand(ctx, repoPath, args...)
	return err
}

// Checkout switches branches or restores files
func (c *Client) Checkout(ctx context.Context, repoPath, branchOrCommit string, options ...CheckoutOption) error {
	args := []string{"checkout"}

	// Apply options
	opts := &CheckoutOptions{}
	for _, option := range options {
		option(opts)
	}

	if opts.CreateBranch {
		args = append(args, "-b")
	}
	if opts.Force {
		args = append(args, "--force")
	}

	args = append(args, branchOrCommit)

	_, err := c.runGitCommand(ctx, repoPath, args...)
	return err
}

// GetCurrentBranch returns the current branch name
func (c *Client) GetCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	output, err := c.runGitCommand(ctx, repoPath, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return output, nil
}

// GetRemoteURL returns the URL of the specified remote
func (c *Client) GetRemoteURL(ctx context.Context, repoPath, remote string) (string, error) {
	if remote == "" {
		remote = "origin"
	}

	output, err := c.runGitCommand(ctx, repoPath, "remote", "get-url", remote)
	if err != nil {
		return "", err
	}
	return output, nil
}
