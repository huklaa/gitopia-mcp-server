//go:build integration

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTier9 exercises context switching: set_active_dao and refresh_user_context.
func TestTier9(t *testing.T) {
	if state.username == "" {
		t.Skip("tier 0 did not complete: no username")
	}

	// Refresh context first to pick up any DAOs created in tier 7
	t.Run("refresh_before_dao_switch", func(t *testing.T) {
		text := callTool(t, "refresh_user_context", nil)
		t.Logf("refresh_user_context response: %s", text)
		assert.Contains(t, text, "refreshed")
	})

	t.Run("set_active_dao", func(t *testing.T) {
		if state.daoName == "" {
			t.Skip("tier 7 did not complete: no DAO")
		}
		text := callTool(t, "set_active_dao", map[string]any{
			"dao_name": state.daoName,
		})
		t.Logf("set_active_dao response: %s", text)
		assert.Contains(t, text, state.daoName)

		// Verify context now shows the DAO as active
		ctxText := callTool(t, "get_user_context", nil)
		t.Logf("get_user_context after set_active_dao: %s", ctxText)
		ctxM := parseQueryData(t, ctxText)
		// active_dao is an object with name, id, address, is_owner
		activeDAO, ok := ctxM["active_dao"].(map[string]any)
		assert.True(t, ok, "active_dao should be an object")
		assert.Equal(t, state.daoName, activeDAO["name"])
	})

	t.Run("clear_active_dao", func(t *testing.T) {
		// Revert to personal account
		text := callTool(t, "set_active_dao", map[string]any{
			"dao_name": "",
		})
		t.Logf("clear_active_dao response: %s", text)
		assert.Contains(t, text, "personal account")

		// Verify no active DAO in context
		ctxText := callTool(t, "get_user_context", nil)
		t.Logf("get_user_context after clear: %s", ctxText)
		ctxM := parseQueryData(t, ctxText)
		assert.Empty(t, ctxM["active_dao"])
	})

	t.Run("refresh_user_context", func(t *testing.T) {
		text := callTool(t, "refresh_user_context", nil)
		t.Logf("refresh_user_context response: %s", text)
		assert.Contains(t, text, "refreshed")

		// Should still have our username
		ctxText := callTool(t, "get_user_context", nil)
		assert.Contains(t, ctxText, state.username)
	})
}
