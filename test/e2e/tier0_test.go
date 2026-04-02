//go:build integration

package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTier0 exercises identity bootstrap: get_user_context, claim_fee_grant,
// and create_user. These must run first because all downstream tiers depend
// on having an initialized wallet identity with an on-chain user.
func TestTier0(t *testing.T) {
	t.Run("get_user_context", func(t *testing.T) {
		text := callTool(t, "get_user_context", nil)
		t.Logf("get_user_context response: %s", text)

		// Response is JSON. Two formats:
		// 1. Auto-generated (mutation): {"status":"success","message":"New wallet auto-generated","address":"gitopia1...","username":"...","fee_grant":"...","user":"..."}
		// 2. Existing user (query):     {"status":"success","message":"User context","data":{"address":"gitopia1...","username":"...","active_dao":"..."}}
		m := tryParseJSON(text)
		require.NotNil(t, m, "get_user_context response should be JSON")

		// Extract address and username from either format
		if data, ok := m["data"].(map[string]any); ok {
			// Query envelope format
			if addr, ok := data["address"].(string); ok {
				state.address = addr
			}
			if user, ok := data["username"].(string); ok && user != "" {
				state.username = user
			}
		} else {
			// Mutation format (auto-generated)
			if addr, ok := m["address"].(string); ok {
				state.address = addr
			}
			if user, ok := m["username"].(string); ok && user != "" {
				state.username = user
			}
		}

		// Fallback: regex for address from raw text
		if state.address == "" {
			addrRe := regexp.MustCompile(`(gitopia1[a-z0-9]{38,})`)
			addrMatch := addrRe.FindString(text)
			require.NotEmpty(t, addrMatch, "should contain a gitopia address")
			state.address = addrMatch
		}

		require.NotEmpty(t, state.address, "should have extracted an address")
		assert.True(t, strings.HasPrefix(state.address, "gitopia1"))
		t.Logf("identity: username=%q address=%q", state.username, state.address)
	})

	t.Run("claim_fee_grant", func(t *testing.T) {
		if state.address == "" {
			t.Skip("no address from get_user_context")
		}
		// claim_fee_grant is best-effort; it may succeed or report already-claimed.
		result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
			Name:      "claim_fee_grant",
			Arguments: map[string]any{"address": state.address},
		})
		require.NoError(t, err, "transport error")
		text := extractText(t, result)
		t.Logf("claim_fee_grant response (isError=%v): %s", result.IsError, text)
		// We don't fail regardless of IsError — already-claimed is fine
	})

	t.Run("create_user_if_needed", func(t *testing.T) {
		if state.address == "" {
			t.Skip("no address from get_user_context")
		}
		if state.username != "" {
			t.Skip("user already exists")
		}
		// Fresh wallet with no on-chain user — create one.
		// Fee grant must have been claimed first (previous subtest).
		username := "e2e-" + state.runID[4:] // e.g. "e2e-1740787200"
		if len(username) > 39 {
			username = username[:39]
		}
		text := callTool(t, "create_user", map[string]any{
			"username": username,
		})
		t.Logf("create_user response: %s", text)
		state.username = username
		chainWritePause()

		// Refresh context so downstream tools see the new user
		_ = callTool(t, "refresh_user_context", nil)
	})
}
