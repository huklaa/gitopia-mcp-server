package handler

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDryRunResult_OutputFormat(t *testing.T) {
	params := map[string]string{"owner": "test", "name": "repo"}
	result, extra, err := DryRunResult("create_issue", params)
	require.NoError(t, err)
	assert.Nil(t, extra)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &parsed))
	assert.Equal(t, "success", parsed["status"])
	assert.Equal(t, true, parsed["dry_run"])
	assert.Equal(t, "create_issue", parsed["tool"])
	assert.NotNil(t, parsed["params"])
}

func TestDryRunResult_ContainsParams(t *testing.T) {
	type testParams struct {
		Owner string `json:"owner"`
		Name  string `json:"name"`
		Title string `json:"title"`
	}
	params := testParams{Owner: "myorg", Name: "myrepo", Title: "Bug fix"}
	result, _, err := DryRunResult("create_issue", params)
	require.NoError(t, err)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "myorg")
	assert.Contains(t, text, "myrepo")
	assert.Contains(t, text, "Bug fix")
	assert.Contains(t, text, "dry_run")
}

func TestDryRunResult_NilParams(t *testing.T) {
	result, _, err := DryRunResult("some_tool", nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(*mcp.TextContent).Text
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &parsed))
	assert.Equal(t, "success", parsed["status"])
	assert.Equal(t, true, parsed["dry_run"])
	assert.Equal(t, "some_tool", parsed["tool"])
	assert.Nil(t, parsed["params"])
}

func TestDryRunResult_ValidJSON(t *testing.T) {
	params := map[string]any{
		"nested": map[string]string{"key": "value"},
		"number": 42,
		"flag":   true,
	}
	result, _, err := DryRunResult("complex_tool", params)
	require.NoError(t, err)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.True(t, json.Valid([]byte(text)))
}
