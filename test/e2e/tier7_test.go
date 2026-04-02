//go:build integration

package e2e

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTier7 exercises DAO operations: create, list_members, submit_proposal,
// vote, get_proposal, exec, and update_members.
func TestTier7(t *testing.T) {
	if state.username == "" || state.address == "" {
		t.Skip("tier 0 did not complete: no identity")
	}

	daoName := state.runID + "-dao"

	t.Run("create_dao", func(t *testing.T) {
		text := callTool(t, "create_dao", map[string]any{
			"name":          daoName,
			"description":   "E2E test DAO from " + state.runID,
			"voting_period": "120",  // 120 seconds to give enough time for voting
			"percentage":    "0.50", // 50% threshold
		})
		t.Logf("create_dao response: %s", text)
		m := parseJSON(t, text)
		assert.Equal(t, "success", m["status"])

		state.daoName = daoName
		chainWritePause()

		// Look up group_id and group_policy_address from the chain
		gid, gpa := lookupDAOGroupInfo(t, daoName)
		state.groupID = gid
		state.groupPolicyAddress = gpa
		t.Logf("DAO created: name=%s group_id=%d group_policy_address=%s",
			state.daoName, state.groupID, state.groupPolicyAddress)
	})

	t.Run("dao_list_members", func(t *testing.T) {
		if state.groupID == 0 {
			t.Skip("create_dao did not produce group_id")
		}
		text := callTool(t, "dao_list_members", map[string]any{
			"group_id": state.groupID,
		})
		t.Logf("dao_list_members response: %s", text)
		arr := parseJSONArray(t, text)
		require.NotEmpty(t, arr, "DAO should have at least one member")
	})

	t.Run("dao_submit_proposal", func(t *testing.T) {
		if state.groupPolicyAddress == "" {
			t.Skip("create_dao did not produce group_policy_address")
		}
		// Do NOT use exec_try here — with a single member at 50% threshold,
		// exec_try would immediately execute the proposal, leaving nothing to vote on.
		text := callTool(t, "dao_submit_proposal", map[string]any{
			"group_policy_address": state.groupPolicyAddress,
			"title":               state.runID + " E2E test proposal",
			"summary":             "Automated proposal from E2E integration test.",
		})
		t.Logf("dao_submit_proposal response: %s", text)
		state.proposalID = extractProposalID(t, text)
		t.Logf("created proposal ID: %d", state.proposalID)
		chainWritePause()
		chainWritePause() // extra pause to avoid sequence mismatch with vote
	})

	t.Run("dao_vote", func(t *testing.T) {
		if state.proposalID == 0 {
			t.Skip("submit_proposal did not complete")
		}
		text := callTool(t, "dao_vote", map[string]any{
			"proposal_id": state.proposalID,
			"option":      "yes",
			"exec_try":    true,
		})
		t.Logf("dao_vote response: %s", text)
		chainWritePause()
	})

	t.Run("dao_get_proposal", func(t *testing.T) {
		if state.proposalID == 0 {
			t.Skip("submit_proposal did not complete")
		}
		text := callTool(t, "dao_get_proposal", map[string]any{
			"proposal_id": state.proposalID,
		})
		t.Logf("dao_get_proposal response: %s", text)
		m := parseQueryData(t, text)
		assert.NotEmpty(t, m["status"])
	})

	t.Run("dao_exec", func(t *testing.T) {
		if state.proposalID == 0 {
			t.Skip("submit_proposal did not complete")
		}
		// Try to execute — may fail if already executed via exec_try on vote.
		// Use error-tolerant call since this is best-effort.
		result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
			Name:      "dao_exec",
			Arguments: map[string]any{"proposal_id": state.proposalID},
		})
		require.NoError(t, err, "transport error")
		text := extractText(t, result)
		t.Logf("dao_exec response (isError=%v): %s", result.IsError, text)
		chainWritePause()
	})

	t.Run("dao_update_members", func(t *testing.T) {
		if state.groupID == 0 {
			t.Skip("create_dao did not produce group_id")
		}
		// Note: The group admin is the group_policy_address, not the user.
		// Direct update_members will fail with "not group admin" unless the
		// user IS the group admin. This test validates the RPC path works.
		result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
			Name: "dao_update_members",
			Arguments: map[string]any{
				"group_id": state.groupID,
				"member_updates": []map[string]any{
					{
						"address":  state.address,
						"weight":   "1",
						"metadata": "updated by E2E",
					},
				},
			},
		})
		require.NoError(t, err, "transport error")
		text := extractText(t, result)
		t.Logf("dao_update_members response (isError=%v): %s", result.IsError, text)
		if result.IsError {
			// Expected: "not group admin" because admin is group_policy_address
			t.Log("NOTE: update_members requires group admin (group_policy_address), not individual user — expected failure")
		}
		chainWritePause()
	})
}
