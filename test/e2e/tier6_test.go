//go:build integration

package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTier6 exercises bounty operations: create, get, list, update, close, delete.
// Uses the smallest possible amount (1 ulore).
// Note: Fee-grant-only wallets have no spendable balance, so create_bounty
// may fail with "insufficient funds". We skip gracefully in that case.
func TestTier6(t *testing.T) {
	if state.repoName == "" || state.issueIID == 0 {
		t.Skip("tier 2/4 did not complete: need repo and issue")
	}

	// Reopen the issue first so we can attach a bounty
	t.Run("reopen_issue_for_bounty", func(t *testing.T) {
		text := callTool(t, "update_issue", map[string]any{
			"owner":        state.repoOwner,
			"name":         state.repoName,
			"issue_iid":    state.issueIID,
			"toggle_state": true,
		})
		t.Logf("reopen issue response: %s", text)
		chainWritePause()
	})

	t.Run("create_bounty", func(t *testing.T) {
		expiry := time.Now().Add(24 * time.Hour).Unix()
		// Use error-tolerant call; fee-grant-only wallets have no spendable balance
		result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
			Name: "create_bounty",
			Arguments: map[string]any{
				"owner":     state.repoOwner,
				"name":      state.repoName,
				"issue_iid": state.issueIID,
				"amount": []map[string]any{
					{"denom": "ulore", "amount": "1"},
				},
				"expiry": expiry,
			},
		})
		require.NoError(t, err, "transport error")
		text := extractText(t, result)
		t.Logf("create_bounty response (isError=%v): %s", result.IsError, text)

		if result.IsError {
			if strings.Contains(text, "insufficient funds") {
				t.Skip("wallet has no spendable balance (fee-grant-only) — skipping bounty tests")
			}
			t.Fatalf("unexpected create_bounty error: %s", text)
		}
		state.bountyID = extractBountyID(t, text)
		t.Logf("created bounty ID: %d", state.bountyID)
		chainWritePause()
	})

	t.Run("get_bounty", func(t *testing.T) {
		if state.bountyID == 0 {
			t.Skip("create_bounty did not complete")
		}
		text := callTool(t, "get_bounty", map[string]any{
			"bounty_id": state.bountyID,
		})
		t.Logf("get_bounty response: %s", text)
		m := parseQueryData(t, text)
		// id may be float64 or string depending on JSON marshaling
		assert.NotNil(t, m["id"])
	})

	t.Run("list_bounties_after_create", func(t *testing.T) {
		text := callTool(t, "list_bounties", nil)
		t.Logf("list_bounties response length: %d", len(text))
		arr := parseJSONArray(t, text)
		assert.NotEmpty(t, arr)
	})

	t.Run("update_bounty", func(t *testing.T) {
		if state.bountyID == 0 {
			t.Skip("create_bounty did not complete")
		}
		newExpiry := time.Now().Add(48 * time.Hour).Unix()
		text := callTool(t, "update_bounty", map[string]any{
			"bounty_id": state.bountyID,
			"expiry":    newExpiry,
		})
		t.Logf("update_bounty response: %s", text)
		chainWritePause()
	})

	t.Run("close_bounty", func(t *testing.T) {
		if state.bountyID == 0 {
			t.Skip("create_bounty did not complete")
		}
		text := callTool(t, "close_bounty", map[string]any{
			"bounty_id": state.bountyID,
		})
		t.Logf("close_bounty response: %s", text)
		chainWritePause()
	})

	t.Run("delete_bounty", func(t *testing.T) {
		if state.bountyID == 0 {
			t.Skip("create_bounty did not complete")
		}
		text := callTool(t, "delete_bounty", map[string]any{
			"bounty_id": state.bountyID,
		})
		t.Logf("delete_bounty response: %s", text)
		chainWritePause()
	})
}
