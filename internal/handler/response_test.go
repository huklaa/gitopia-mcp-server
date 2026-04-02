package handler

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseResponse extracts the JSON text from a CallToolResult and unmarshals it.
func parseResponse(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	tc, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &m), "response is not valid JSON: %s", tc.Text)
	return m
}

func TestToolSuccess(t *testing.T) {
	result, meta, err := toolSuccess("Created repo", map[string]any{
		"url": "https://gitopia.com/alice/repo",
	})
	require.NoError(t, err)
	assert.Nil(t, meta)
	assert.False(t, result.IsError)

	m := parseResponse(t, result)
	assert.Equal(t, "success", m["status"])
	assert.Equal(t, "Created repo", m["message"])
	assert.Equal(t, "https://gitopia.com/alice/repo", m["url"])
}

func TestToolSuccessNoExtra(t *testing.T) {
	result, _, err := toolSuccess("Done", nil)
	require.NoError(t, err)

	m := parseResponse(t, result)
	assert.Equal(t, "success", m["status"])
	assert.Equal(t, "Done", m["message"])
}

func TestToolSuccessData(t *testing.T) {
	data := []map[string]any{
		{"name": "repo1", "id": float64(1)},
		{"name": "repo2", "id": float64(2)},
	}
	result, meta, err := toolSuccessData("Found 2 repos", data)
	require.NoError(t, err)
	assert.Nil(t, meta)
	assert.False(t, result.IsError)

	m := parseResponse(t, result)
	assert.Equal(t, "success", m["status"])
	assert.Equal(t, "Found 2 repos", m["message"])
	items, ok := m["data"].([]any)
	require.True(t, ok)
	assert.Len(t, items, 2)
}

func TestToolErrorf(t *testing.T) {
	result, meta, err := toolErrorf(ErrNotFound, "Repo %q not found", "alice/repo")
	require.NoError(t, err)
	assert.Nil(t, meta)
	assert.True(t, result.IsError)

	m := parseResponse(t, result)
	assert.Equal(t, "error", m["status"])
	assert.Equal(t, `Repo "alice/repo" not found`, m["message"])
	assert.Equal(t, "NOT_FOUND", m["code"])
}

func TestToolErrorfCodes(t *testing.T) {
	codes := []ErrorCode{
		ErrClientUnavailable, ErrAuthFailed, ErrRateLimited,
		ErrValidation, ErrNotFound, ErrConflict,
		ErrNotRepository, ErrPathTraversal, ErrChainTx,
		ErrGitOperation, ErrFileSystem, ErrApproval, ErrInternal,
	}
	for _, code := range codes {
		result, _, err := toolErrorf(code, "test")
		require.NoError(t, err)
		assert.True(t, result.IsError)

		m := parseResponse(t, result)
		assert.Equal(t, string(code), m["code"])
	}
}

func TestJsonResult(t *testing.T) {
	result, _, err := jsonResult(map[string]any{
		"status":  "success",
		"message": "hello",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	m := parseResponse(t, result)
	assert.Equal(t, "success", m["status"])
}
