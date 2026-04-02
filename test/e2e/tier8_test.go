//go:build integration

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTier8 exercises batch_execute and the approval mode flow
// (confirm_transaction, reject_transaction, list_pending_transactions).
func TestTier8(t *testing.T) {
	if state.repoName == "" || state.issueIID == 0 {
		t.Skip("tier 2/4 did not complete: need repo and issue")
	}

	t.Run("batch_execute", func(t *testing.T) {
		if state.repoID == 0 {
			t.Skip("no repo_id available")
		}
		// Batch two comments on the issue using repo_id (batch operations
		// use numeric repo_id, not owner/name).
		text := callTool(t, "batch_execute", map[string]any{
			"operations": []map[string]any{
				{
					"tool": "comment_on_issue",
					"params": map[string]any{
						"repo_id":   state.repoID,
						"issue_iid": state.issueIID,
						"body":      "Batch comment 1 from " + state.runID,
					},
				},
				{
					"tool": "comment_on_issue",
					"params": map[string]any{
						"repo_id":   state.repoID,
						"issue_iid": state.issueIID,
						"body":      "Batch comment 2 from " + state.runID,
					},
				},
			},
		})
		t.Logf("batch_execute response: %s", text)
		chainWritePause()
	})

	// ---- Approval Mode Tests ----
	// These use a second in-process server with ApprovalMode=true.
	// Only tools that go through signOrHold support approval mode:
	// dao_submit_proposal, dao_vote, dao_exec,
	// dao_update_members, batch_execute.

	t.Run("approval_mode", func(t *testing.T) {
		if state.approvalSession == nil {
			t.Skip("approval mode server not available")
		}
		if state.groupPolicyAddress == "" {
			t.Skip("tier 7 did not complete: no group_policy_address for approval test")
		}
		session := state.approvalSession

		t.Run("list_pending_empty", func(t *testing.T) {
			text := callToolOn(t, session, "list_pending_transactions", nil)
			t.Logf("list_pending (empty): %s", text)
			// Should return empty list or message about no pending transactions
		})

		t.Run("hold_then_reject", func(t *testing.T) {
			// Use dao_submit_proposal which goes through signOrHold
			text := callToolOn(t, session, "dao_submit_proposal", map[string]any{
				"group_policy_address": state.groupPolicyAddress,
				"title":               state.runID + " approval reject test",
				"summary":             "This proposal should be rejected by approval mode test.",
			})
			t.Logf("held proposal response: %s", text)

			// Extract pending_id from the approval preview JSON
			pendingID := extractPendingID(t, text)
			require.NotEmpty(t, pendingID)

			// List pending — should have our tx
			listText := callToolOn(t, session, "list_pending_transactions", nil)
			t.Logf("list_pending (with tx): %s", listText)
			assert.Contains(t, listText, pendingID)

			// Reject the transaction
			rejectText := callToolOn(t, session, "reject_transaction", map[string]any{
				"pending_id": pendingID,
			})
			t.Logf("reject_transaction response: %s", rejectText)

			// List pending — should be empty again
			listText2 := callToolOn(t, session, "list_pending_transactions", nil)
			t.Logf("list_pending (after reject): %s", listText2)
			assert.NotContains(t, listText2, pendingID)
		})

		t.Run("hold_then_confirm", func(t *testing.T) {
			// Create a new held transaction
			text := callToolOn(t, session, "dao_submit_proposal", map[string]any{
				"group_policy_address": state.groupPolicyAddress,
				"title":               state.runID + " approval confirm test",
				"summary":             "This proposal should be confirmed by approval mode test.",
			})
			t.Logf("held proposal for confirm: %s", text)

			pendingID := extractPendingID(t, text)
			require.NotEmpty(t, pendingID)

			// Confirm the transaction (actually broadcasts)
			confirmText := callToolOn(t, session, "confirm_transaction", map[string]any{
				"pending_id": pendingID,
			})
			t.Logf("confirm_transaction response: %s", confirmText)
			// Should indicate success (broadcast)
			assert.True(t,
				strings.Contains(strings.ToLower(confirmText), "success") ||
					strings.Contains(strings.ToLower(confirmText), "broadcast") ||
					strings.Contains(strings.ToLower(confirmText), "confirmed") ||
					strings.Contains(confirmText, "proposal"),
				"confirm should indicate successful broadcast, got: %s", confirmText)
			chainWritePause()
		})
	})
}
