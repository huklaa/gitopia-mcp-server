//go:build integration

package e2e

import (
	"encoding/json"
	"regexp"
	"strconv"
	"testing"
	"time"

	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// callTool calls a tool via the MCP client session and returns the text content.
// It asserts that there is no transport error and IsError is false.
func callTool(t *testing.T, name string, args map[string]any) string {
	t.Helper()
	result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	require.NoError(t, err, "transport error calling %s", name)
	text := extractText(t, result)
	require.False(t, result.IsError, "tool %s returned error: %s", name, text)
	return text
}

// callToolExpectError calls a tool and asserts that IsError is true.
func callToolExpectError(t *testing.T, name string, args map[string]any) string {
	t.Helper()
	result, err := state.session.CallTool(state.ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	require.NoError(t, err, "transport error calling %s", name)
	text := extractText(t, result)
	require.True(t, result.IsError, "expected tool %s to return IsError=true but got: %s", name, text)
	return text
}

// callToolOn calls a tool on a specific session (e.g. approval session).
func callToolOn(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	result, err := session.CallTool(state.ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	require.NoError(t, err, "transport error calling %s", name)
	text := extractText(t, result)
	require.False(t, result.IsError, "tool %s returned error: %s", name, text)
	return text
}

// callToolOnExpectError calls a tool on a specific session and expects IsError.
func callToolOnExpectError(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	result, err := session.CallTool(state.ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	require.NoError(t, err, "transport error calling %s", name)
	text := extractText(t, result)
	require.True(t, result.IsError, "expected tool %s to return IsError=true but got: %s", name, text)
	return text
}

// extractText extracts the first TextContent from a CallToolResult.
func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, result, "result is nil")
	require.NotEmpty(t, result.Content, "result has no content")
	tc, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "first content is not TextContent: %T", result.Content[0])
	return tc.Text
}

// parseJSON unmarshals a JSON object string.
func parseJSON(t *testing.T, text string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &m), "failed to parse JSON object: %s", text)
	return m
}

// parseQueryData parses a JSON envelope and extracts the "data" field as a map.
// Handles the standard query response format: {"status":"success","message":"...","data":{...}}
func parseQueryData(t *testing.T, text string) map[string]any {
	t.Helper()
	m := parseJSON(t, text)
	data, ok := m["data"]
	require.True(t, ok, "response has no 'data' field: %s", text)
	obj, ok := data.(map[string]any)
	require.True(t, ok, "response 'data' is not an object: %T", data)
	return obj
}

// parseJSONArray parses a JSON response and returns an array of objects.
// Handles both raw JSON arrays and enveloped responses: {"status":"success","data":[...]}
func parseJSONArray(t *testing.T, text string) []map[string]any {
	t.Helper()
	// Try envelope format first
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err == nil {
		if data, ok := m["data"]; ok {
			if arr, ok := data.([]any); ok {
				result := make([]map[string]any, 0, len(arr))
				for _, item := range arr {
					if obj, ok := item.(map[string]any); ok {
						result = append(result, obj)
					}
				}
				return result
			}
		}
	}
	// Fallback: raw JSON array
	var arr []map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &arr), "failed to parse JSON array: %s", text)
	return arr
}

// tryParseJSON tries to parse as JSON object; returns nil if not JSON.
func tryParseJSON(text string) map[string]any {
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		return nil
	}
	return m
}

// extractIssueIID extracts an issue IID from response text.
// Looks for patterns like "iid":N or "number":N or "issue_iid":N or "Issue #N".
func extractIssueIID(t *testing.T, text string) uint64 {
	t.Helper()
	// Try JSON first
	if m := tryParseJSON(text); m != nil {
		for _, key := range []string{"iid", "number", "issue_iid"} {
			if v, ok := m[key]; ok {
				return toUint64(t, v)
			}
		}
	}
	// Regex fallback — also match issues/N in URLs
	re := regexp.MustCompile(`(?:iid|number|Issue #|issue_iid|issues/)["\s:]*(\d+)`)
	match := re.FindStringSubmatch(text)
	require.NotEmpty(t, match, "could not extract issue IID from: %s", text)
	n, err := strconv.ParseUint(match[1], 10, 64)
	require.NoError(t, err)
	return n
}

// extractBountyID extracts a bounty ID from response text.
func extractBountyID(t *testing.T, text string) uint64 {
	t.Helper()
	if m := tryParseJSON(text); m != nil {
		for _, key := range []string{"bounty_id", "id"} {
			if v, ok := m[key]; ok {
				return toUint64(t, v)
			}
		}
	}
	re := regexp.MustCompile(`(?:bounty_id|bounty ID|"id")["\s:]*(\d+)`)
	match := re.FindStringSubmatch(text)
	require.NotEmpty(t, match, "could not extract bounty ID from: %s", text)
	n, err := strconv.ParseUint(match[1], 10, 64)
	require.NoError(t, err)
	return n
}

// extractPRNumber extracts a PR number from response text.
func extractPRNumber(t *testing.T, text string) uint64 {
	t.Helper()
	if m := tryParseJSON(text); m != nil {
		for _, key := range []string{"pull_iid", "pr_number", "iid", "number"} {
			if v, ok := m[key]; ok {
				return toUint64(t, v)
			}
		}
	}
	re := regexp.MustCompile(`(?:pull_iid|pr_number|PR #|number|iid|pulls/)["\s:]*(\d+)`)
	match := re.FindStringSubmatch(text)
	require.NotEmpty(t, match, "could not extract PR number from: %s", text)
	n, err := strconv.ParseUint(match[1], 10, 64)
	require.NoError(t, err)
	return n
}

// extractDAOInfo extracts DAO name, group_id, and group_policy_address from
// a create_dao or get_user_context response.
func extractDAOInfo(t *testing.T, text string) (name string, groupID uint64, groupPolicyAddress string) {
	t.Helper()
	m := parseJSON(t, text)
	if v, ok := m["name"]; ok {
		name, _ = v.(string)
	}
	if v, ok := m["group_id"]; ok {
		groupID = toUint64(t, v)
	}
	if v, ok := m["group_policy_address"]; ok {
		groupPolicyAddress, _ = v.(string)
	}
	return
}

// extractProposalID extracts a proposal ID from response text.
func extractProposalID(t *testing.T, text string) uint64 {
	t.Helper()
	if m := tryParseJSON(text); m != nil {
		for _, key := range []string{"proposal_id", "id"} {
			if v, ok := m[key]; ok {
				return toUint64(t, v)
			}
		}
	}
	re := regexp.MustCompile(`(?:proposal_id|proposal ID|Proposal ID|"id")["\s:]*(\d+)`)
	match := re.FindStringSubmatch(text)
	require.NotEmpty(t, match, "could not extract proposal ID from: %s", text)
	n, err := strconv.ParseUint(match[1], 10, 64)
	require.NoError(t, err)
	return n
}

// extractPendingID extracts a pending transaction ID from approval mode response.
func extractPendingID(t *testing.T, text string) string {
	t.Helper()
	if m := tryParseJSON(text); m != nil {
		if v, ok := m["pending_id"]; ok {
			s, _ := v.(string)
			require.NotEmpty(t, s, "pending_id is empty")
			return s
		}
	}
	re := regexp.MustCompile(`pending_id["\s:]*"?([a-f0-9]+)"?`)
	match := re.FindStringSubmatch(text)
	require.NotEmpty(t, match, "could not extract pending_id from: %s", text)
	return match[1]
}

// toUint64 converts a JSON number (float64) or string to uint64.
func toUint64(t *testing.T, v any) uint64 {
	t.Helper()
	switch val := v.(type) {
	case float64:
		return uint64(val)
	case json.Number:
		n, err := val.Int64()
		require.NoError(t, err)
		return uint64(n)
	case string:
		n, err := strconv.ParseUint(val, 10, 64)
		require.NoError(t, err)
		return n
	default:
		t.Fatalf("cannot convert %T to uint64", v)
		return 0
	}
}

// chainWritePause sleeps between consecutive chain-write operations to avoid
// account sequence mismatches.
func chainWritePause() {
	time.Sleep(3 * time.Second)
}

// lookupDAOGroupInfo queries the chain for a DAO's group_id and group_policy_address.
// Uses the gitopia query client to resolve DAO name -> group_id -> group policies.
func lookupDAOGroupInfo(t *testing.T, daoName string) (groupID uint64, groupPolicyAddress string) {
	t.Helper()

	q := gitopiatypes.NewQueryClient(state.gClient.GetConn())
	resp, err := q.Dao(state.ctx, &gitopiatypes.QueryGetDaoRequest{Id: daoName})
	require.NoError(t, err, "failed to query DAO %s", daoName)
	require.NotNil(t, resp)

	groupID = resp.Dao.GroupId
	require.NotZero(t, groupID, "DAO %s has no group_id", daoName)

	// Look up group policies for this group
	policies, err := state.gClient.GetGroupPoliciesByGroup(state.ctx, groupID)
	require.NoError(t, err, "failed to get group policies for group %d", groupID)
	require.NotEmpty(t, policies, "no group policies for group %d", groupID)

	groupPolicyAddress = policies[0].Address
	return
}
