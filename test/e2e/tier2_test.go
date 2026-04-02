//go:build integration

package e2e

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTier2 exercises repository creation via bootstrap_repo, then verifies
// with get_repo, list_branches, and get_file_contents.
func TestTier2(t *testing.T) {
	if state.username == "" {
		t.Skip("tier 0 did not complete: no username")
	}

	repoName := state.runID + "-repo"

	t.Run("bootstrap_repo", func(t *testing.T) {
		// bootstrap_repo creates the repo on-chain, initializes locally, and pushes.
		// The push may fail with "packfile doesn't exist" — a known chain-side race
		// condition where the git server hasn't synced the packfile storage.
		// We accept partial success: repo created on-chain + local init.
		result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
			Name: "bootstrap_repo",
			Arguments: map[string]any{
				"owner_id":         state.username,
				"name":             repoName,
				"description":      "E2E test repository created by " + state.runID,
				"local_path":       repoName,
				"create_readme":    true,
				"create_gitignore": false,
				"initial_branch":   "main",
			},
		})
		require.NoError(t, err, "transport error")
		text := extractText(t, result)
		t.Logf("bootstrap_repo response: %s", text)

		// Always record the repo — the on-chain create succeeds even if push fails
		state.repoName = repoName
		state.repoOwner = state.username

		// Check for success: JSON status field or legacy text
		m := tryParseJSON(text)
		isSuccess := (m != nil && m["status"] == "success") || strings.Contains(text, "success")
		if isSuccess && !strings.Contains(text, "packfile") {
			state.repoPushed = true
			// Wait for chain to propagate the push before querying branches/files
			chainWritePause()
			chainWritePause() // extra wait for git server sync
		} else if strings.Contains(text, "packfile") {
			t.Log("KNOWN ISSUE: push failed with 'packfile doesn't exist' — chain-side race condition")
			t.Log("Repo was created on-chain; push-dependent subtests will be skipped")
		} else {
			t.Errorf("unexpected bootstrap_repo failure: %s", text)
		}
	})

	t.Run("get_repo", func(t *testing.T) {
		if state.repoName == "" {
			t.Skip("bootstrap_repo did not complete")
		}
		text := callTool(t, "get_repo", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("get_repo response: %s", text)
		m := parseQueryData(t, text)
		assert.Equal(t, state.repoName, m["name"])
		if v, ok := m["id"]; ok {
			state.repoID = toUint64(t, v)
			t.Logf("repo ID: %d", state.repoID)
		}
	})

	t.Run("list_branches", func(t *testing.T) {
		if !state.repoPushed {
			t.Skip("push did not succeed — no branches to list")
		}
		text := callTool(t, "list_branches", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("list_branches response: %s", text)
		arr := parseJSONArray(t, text)
		require.NotEmpty(t, arr, "should have at least one branch")
		found := false
		for _, b := range arr {
			if b["name"] == "main" {
				found = true
				break
			}
		}
		assert.True(t, found, "main branch should exist")
	})

	t.Run("get_file_contents", func(t *testing.T) {
		if !state.repoPushed {
			t.Skip("push did not succeed — no files to read")
		}
		text := callTool(t, "get_file_contents", map[string]any{
			"owner":  state.repoOwner,
			"name":   state.repoName,
			"branch": "main",
			"path":   "README.md",
		})
		t.Logf("get_file_contents response length: %d", len(text))
		assert.Contains(t, text, state.repoName, "README should mention repo name")
	})
}
