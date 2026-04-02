//go:build integration

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTier1 exercises read-only queries that do not require a repo created by
// this run. Uses the well-known "gitopia" user which always has repos.
func TestTier1(t *testing.T) {
	t.Run("list_repos_known_user", func(t *testing.T) {
		text := callTool(t, "list_repos", map[string]any{
			"owner": "gitopia",
		})
		t.Logf("list_repos(gitopia) response length: %d", len(text))
		arr := parseJSONArray(t, text)
		require.NotEmpty(t, arr, "gitopia user should have at least one repo")
		// Check first repo has expected fields
		first := arr[0]
		assert.NotEmpty(t, first["name"], "repo should have a name")
	})

	t.Run("list_bounties", func(t *testing.T) {
		text := callTool(t, "list_bounties", nil)
		t.Logf("list_bounties response length: %d", len(text))
		// May be empty array, but should parse
		_ = parseJSONArray(t, text)
	})
}
