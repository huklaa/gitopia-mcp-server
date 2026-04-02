//go:build integration

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTier4 exercises issue operations: create, get, list, comment, update/close.
func TestTier4(t *testing.T) {
	if state.repoName == "" {
		t.Skip("tier 2 did not complete: no repo")
	}

	t.Run("create_issue", func(t *testing.T) {
		text := callTool(t, "create_issue", map[string]any{
			"owner":       state.repoOwner,
			"name":        state.repoName,
			"title":       state.runID + " E2E test issue",
			"description": "This issue was created by the E2E integration test suite.",
		})
		t.Logf("create_issue response: %s", text)
		state.issueIID = extractIssueIID(t, text)
		t.Logf("created issue IID: %d", state.issueIID)
		chainWritePause()
	})

	t.Run("get_issue", func(t *testing.T) {
		if state.issueIID == 0 {
			t.Skip("create_issue did not complete")
		}
		text := callTool(t, "get_issue", map[string]any{
			"owner":     state.repoOwner,
			"name":      state.repoName,
			"issue_iid": state.issueIID,
		})
		t.Logf("get_issue response: %s", text)
		m := parseQueryData(t, text)
		// Title contains the run ID
		title, _ := m["title"].(string)
		assert.Contains(t, title, state.runID)
	})

	t.Run("list_issues", func(t *testing.T) {
		if state.issueIID == 0 {
			t.Skip("create_issue did not complete")
		}
		text := callTool(t, "list_issues", map[string]any{
			"owner": state.repoOwner,
			"name":  state.repoName,
		})
		t.Logf("list_issues response length: %d", len(text))
		arr := parseJSONArray(t, text)
		require.NotEmpty(t, arr, "should have at least one issue")
	})

	t.Run("comment_on_issue", func(t *testing.T) {
		if state.issueIID == 0 {
			t.Skip("create_issue did not complete")
		}
		text := callTool(t, "comment_on_issue", map[string]any{
			"owner":     state.repoOwner,
			"name":      state.repoName,
			"issue_iid": state.issueIID,
			"body":      "E2E automated comment from " + state.runID,
		})
		t.Logf("comment_on_issue response: %s", text)
		chainWritePause()
	})

	t.Run("update_issue_close", func(t *testing.T) {
		if state.issueIID == 0 {
			t.Skip("create_issue did not complete")
		}
		text := callTool(t, "update_issue", map[string]any{
			"owner":         state.repoOwner,
			"name":          state.repoName,
			"issue_iid":     state.issueIID,
			"toggle_state":  true,
			"state_comment": "Closing from E2E test",
		})
		t.Logf("update_issue(close) response: %s", text)
		chainWritePause()
	})
}
