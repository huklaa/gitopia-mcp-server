//go:build integration

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTier3 exercises git push and file content operations.
// Note: write_file, read_file, git_status, git_add, git_commit, and search_code
// were removed in v0.2 (redundant with MCP client built-ins).
// This tier now tests commit_and_push_changes and get_file_contents.
func TestTier3(t *testing.T) {
	if state.repoName == "" {
		t.Skip("tier 2 did not complete: no repo")
	}

	repoPath := state.repoName

	t.Run("commit_and_push_changes", func(t *testing.T) {
		// commit_and_push_changes stages all, commits, and pushes in one step
		text := callTool(t, "commit_and_push_changes", map[string]any{
			"repo_path":      repoPath,
			"commit_message": "E2E: test commit via commit_and_push_changes",
			"files":          []string{},
		})
		t.Logf("commit_and_push_changes response: %s", text)
		// May succeed or report "nothing to commit" — both are valid
		assert.NotEmpty(t, text)
		chainWritePause()
	})

	t.Run("get_file_contents_remote", func(t *testing.T) {
		// Read a file from the remote repo without cloning
		text := callTool(t, "get_file_contents", map[string]any{
			"owner":  state.repoOwner,
			"name":   state.repoName,
			"branch": "main",
			"path":   "README.md",
		})
		t.Logf("get_file_contents response: %s", text)
		assert.Contains(t, text, state.repoName, "README should mention repo name")
	})
}
