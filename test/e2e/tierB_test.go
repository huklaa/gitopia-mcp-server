//go:build integration

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTierB exercises the 7 new tools: tags, commits, releases, labels.
// Labels are on-chain metadata and don't require a push.
// Tags, commits, and releases query the git server and require a pushed repo.
func TestTierB(t *testing.T) {
	if state.repoName == "" {
		t.Skip("tier 2 did not complete: no repo")
	}

	// ---- Label lifecycle (no push dependency) ----

	t.Run("create_label", func(t *testing.T) {
		text := callTool(t, "create_label", map[string]any{
			"owner":       state.repoOwner,
			"name":        state.repoName,
			"label_name":  "bug-" + state.runID,
			"color":       "FF0000",
			"description": "E2E test label",
		})
		t.Logf("create_label response: %s", text)
		m := parseJSON(t, text)
		require.Equal(t, "success", m["status"])
		state.labelID = toUint64(t, m["label_id"])
		require.NotZero(t, state.labelID, "label_id should be non-zero")
		chainWritePause()
	})

	t.Run("list_labels", func(t *testing.T) {
		if state.labelID == 0 {
			t.Skip("label not created")
		}
		text := callTool(t, "list_labels", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("list_labels response: %s", text)
		m := parseJSON(t, text)
		data, ok := m["data"].([]any)
		require.True(t, ok, "data should be an array")
		require.NotEmpty(t, data, "should have at least 1 label")

		// Find our label
		found := false
		for _, item := range data {
			label, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if name, _ := label["name"].(string); name == "bug-"+state.runID {
				found = true
				assert.Equal(t, "FF0000", label["color"])
				assert.Equal(t, "E2E test label", label["description"])
			}
		}
		assert.True(t, found, "our label should be in the list")
	})

	t.Run("delete_label", func(t *testing.T) {
		if state.labelID == 0 {
			t.Skip("label not created")
		}
		text := callTool(t, "delete_label", map[string]any{
			"owner":    state.repoOwner,
			"name":     state.repoName,
			"label_id": state.labelID,
		})
		t.Logf("delete_label response: %s", text)
		m := parseJSON(t, text)
		require.Equal(t, "success", m["status"])
		chainWritePause()
	})

	t.Run("list_labels_after_delete", func(t *testing.T) {
		if state.labelID == 0 {
			t.Skip("label not created")
		}
		text := callTool(t, "list_labels", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("list_labels after delete: %s", text)
		m := parseJSON(t, text)
		data, ok := m["data"].([]any)
		require.True(t, ok, "data should be an array")

		// Verify our label is gone
		for _, item := range data {
			label, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name, _ := label["name"].(string)
			assert.NotEqual(t, "bug-"+state.runID, name, "deleted label should not appear")
		}
	})

	// ---- Tags, commits, releases (require push) ----

	t.Run("list_tags", func(t *testing.T) {
		if !state.repoPushed {
			t.Skip("push did not succeed")
		}
		text := callTool(t, "list_tags", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("list_tags response: %s", text)
		m := parseJSON(t, text)
		require.Equal(t, "success", m["status"])
		// New repo may have no tags — that's fine, just verify the response is valid
	})

	t.Run("list_commits_default_branch", func(t *testing.T) {
		if !state.repoPushed {
			t.Skip("push did not succeed")
		}
		// Omit branch param to test DefaultBranch lookup
		text := callTool(t, "list_commits", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("list_commits (default branch) response: %s", text)
		m := parseJSON(t, text)
		require.Equal(t, "success", m["status"])
		data, ok := m["data"].([]any)
		require.True(t, ok, "data should be an array")
		require.NotEmpty(t, data, "pushed repo should have at least 1 commit")

		// Verify commit has expected fields
		commit, ok := data[0].(map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, commit["id"], "commit should have an id")
		assert.NotEmpty(t, commit["title"], "commit should have a title")
		assert.NotNil(t, commit["author"], "commit should have an author")
	})

	t.Run("list_commits_explicit_branch", func(t *testing.T) {
		if !state.repoPushed {
			t.Skip("push did not succeed")
		}
		text := callTool(t, "list_commits", map[string]any{
			"owner":  state.repoOwner,
			"name":   state.repoName,
			"branch": "main",
			"limit":  2,
		})
		t.Logf("list_commits (explicit main) response: %s", text)
		m := parseJSON(t, text)
		require.Equal(t, "success", m["status"])
	})

	t.Run("list_releases", func(t *testing.T) {
		if !state.repoPushed {
			t.Skip("push did not succeed")
		}
		text := callTool(t, "list_releases", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("list_releases response: %s", text)
		m := parseJSON(t, text)
		require.Equal(t, "success", m["status"])
		// New repo has no releases — verify response is valid
	})
}
