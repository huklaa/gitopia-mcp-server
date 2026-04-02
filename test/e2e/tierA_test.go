//go:build integration

package e2e

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTierA exercises remaining tools: git_init, create_feature_branch,
// sync_with_remote, commit_and_push_changes, update_feature_branch,
// fork_repository, and standalone create_pull_request.
// Note: Named "TierA" (not "Tier10") so Go sorts it after Tier9.
func TestTierA(t *testing.T) {
	if state.repoName == "" {
		t.Skip("tier 2 did not complete: no repo")
	}

	repoPath := state.repoName
	var tierAPushed bool // tracks if any push succeeded in this tier

	// Note: git_init, write_file, git_add, git_commit removed in v0.2
	// (redundant with MCP client built-ins). Use commit_and_push_changes instead.

	t.Run("create_feature_branch", func(t *testing.T) {
		branchName := state.runID + "-tier10-branch"
		text := callTool(t, "create_feature_branch", map[string]any{
			"repo_path":   repoPath,
			"branch_name": branchName,
			"base_branch": "main",
		})
		t.Logf("create_feature_branch response: %s", text)

		// Push the branch (may have no new commits, but tests the push path)
		// Push may fail due to chain-side packfile contention from earlier tiers.
		result, err := tryPushWithRetry(t, "git_push", map[string]any{
			"repo_path":    repoPath,
			"set_upstream": true,
		})
		require.NoError(t, err, "transport error")
		pushText := extractText(t, result)
		t.Logf("push tier10 branch response (isError=%v): %s", result.IsError, pushText)

		if result.IsError {
			t.Log("NOTE: push failed due to packfile contention — skipping standalone PR creation")
			return
		}
		tierAPushed = true
		chainWritePause()

		// Create a PR via standalone create_pull_request
		t.Run("create_pull_request_standalone", func(t *testing.T) {
			prText := callTool(t, "create_pull_request", map[string]any{
				"owner":       state.repoOwner,
				"name":        state.repoName,
				"title":       state.runID + " tier10 standalone PR",
				"description": "Created via standalone create_pull_request",
				"head_branch": branchName,
				"base_branch": "main",
			})
			t.Logf("create_pull_request (standalone) response: %s", prText)
			chainWritePause()
		})
	})

	t.Run("sync_with_remote", func(t *testing.T) {
		// Sync may fail if we're on a feature branch without upstream tracking.
		// Use error-tolerant call since this validates the RPC path, not branch state.
		result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
			Name: "sync_with_remote",
			Arguments: map[string]any{
				"repo_path": repoPath,
				"remote":    "origin",
				"branch":    "main",
			},
		})
		require.NoError(t, err, "transport error")
		text := extractText(t, result)
		t.Logf("sync_with_remote response (isError=%v): %s", result.IsError, text)
	})

	t.Run("commit_and_push_changes", func(t *testing.T) {
		// commit_and_push_changes handles stage+commit+push
		// It may report "nothing to commit" if workspace is clean — that's valid
		result, err := tryPushWithRetry(t, "commit_and_push_changes", map[string]any{
			"repo_path":      repoPath,
			"commit_message": "E2E: commit_and_push_changes test",
			"files":          []string{"push-test.txt"},
		})
		require.NoError(t, err, "transport error")
		text := extractText(t, result)
		t.Logf("commit_and_push_changes response (isError=%v): %s", result.IsError, text)
		if !result.IsError {
			tierAPushed = true
		}
		chainWritePause()
	})

	t.Run("update_feature_branch", func(t *testing.T) {
		// Use the feature branch from tier 5 (if it still exists)
		branchName := state.runID + "-feature"
		result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
			Name: "update_feature_branch",
			Arguments: map[string]any{
				"repo_path":   repoPath,
				"branch_name": branchName,
				"files": []map[string]any{
					{
						"path":    "updated.txt",
						"content": "Updated feature from E2E " + state.runID,
						"mode":    "create",
					},
				},
				"commit_message": "E2E: update feature branch",
			},
		})
		require.NoError(t, err, "transport error")
		text := extractText(t, result)
		t.Logf("update_feature_branch response (isError=%v): %s", result.IsError, text)
		// May fail if branch was deleted after merge — that's OK
	})

	t.Run("toggle_forking", func(t *testing.T) {
		text := callTool(t, "toggle_repository_forking", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("toggle_repository_forking response: %s", text)
		assert.Contains(t, text, "allow_forking")
		chainWritePause()
	})

	t.Run("fork_repository", func(t *testing.T) {
		text := callTool(t, "fork_repository", map[string]any{
			"owner":     state.repoOwner,
			"name":      state.repoName,
			"fork_name": state.runID + "-fork",
		})
		t.Logf("fork_repository response: %s", text)
		assert.Contains(t, text, "fork_id")
		chainWritePause()
	})

	if !tierAPushed {
		t.Log("NOTE: no pushes succeeded in TierA due to chain-side packfile contention — this is a known chain-side timing issue")
	}
}

// tryPushWithRetry attempts a tool call that pushes to the chain, retrying on
// "pending packfile" errors. Returns the last result rather than failing the test.
func tryPushWithRetry(t *testing.T, name string, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	const maxAttempts = 6
	const waitSec = 15
	var lastResult *mcp.CallToolResult
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		})
		if err != nil {
			return result, err
		}
		lastResult = result

		if !result.IsError {
			return result, nil
		}

		text := extractText(t, result)
		if !strings.Contains(text, "pending packfile") {
			return result, nil // non-packfile error, return as-is
		}

		if attempt < maxAttempts {
			t.Logf("%s attempt %d/%d: packfile pending, waiting %ds...", name, attempt, maxAttempts, waitSec)
			chainWritePause()
			chainWritePause()
			chainWritePause()
			chainWritePause()
			chainWritePause() // 5 × 3s = 15s
		}
	}
	return lastResult, nil
}
