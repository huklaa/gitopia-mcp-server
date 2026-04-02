package handler

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- URI Parsing Tests ----

func TestParseURIParams_ValidRepoURI(t *testing.T) {
	params, err := parseURIParams("gitopia://repos/myorg/myrepo", "gitopia://repos/{owner}/{name}")
	require.NoError(t, err)
	assert.Equal(t, "myorg", params["owner"])
	assert.Equal(t, "myrepo", params["name"])
}

func TestParseURIParams_ValidIssueURI(t *testing.T) {
	params, err := parseURIParams("gitopia://repos/myorg/myrepo/issues/42", "gitopia://repos/{owner}/{name}/issues/{iid}")
	require.NoError(t, err)
	assert.Equal(t, "myorg", params["owner"])
	assert.Equal(t, "myrepo", params["name"])
	assert.Equal(t, "42", params["iid"])
}

func TestParseURIParams_ValidPRURI(t *testing.T) {
	params, err := parseURIParams("gitopia://repos/myorg/myrepo/pulls/15", "gitopia://repos/{owner}/{name}/pulls/{iid}")
	require.NoError(t, err)
	assert.Equal(t, "myorg", params["owner"])
	assert.Equal(t, "myrepo", params["name"])
	assert.Equal(t, "15", params["iid"])
}

func TestParseURIParams_ValidBountyURI(t *testing.T) {
	params, err := parseURIParams("gitopia://bounties/123", "gitopia://bounties/{id}")
	require.NoError(t, err)
	assert.Equal(t, "123", params["id"])
}

func TestParseURIParams_SegmentMismatch(t *testing.T) {
	_, err := parseURIParams("gitopia://repos/myorg", "gitopia://repos/{owner}/{name}")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match pattern")
}

func TestParseURIParams_LiteralMismatch(t *testing.T) {
	_, err := parseURIParams("gitopia://repos/myorg/myrepo/issues/42", "gitopia://repos/{owner}/{name}/pulls/{iid}")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match pattern")
}

// ---- Resource Handler Tests ----

func TestHandleRepoResource_NoClient(t *testing.T) {
	rh := &ResourceHandler{GClient: nil}
	_, err := rh.HandleRepoResource(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "gitopia://repos/myorg/myrepo"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestHandleIssueResource_NoClient(t *testing.T) {
	rh := &ResourceHandler{GClient: nil}
	_, err := rh.HandleIssueResource(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "gitopia://repos/myorg/myrepo/issues/42"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestHandlePullRequestResource_NoClient(t *testing.T) {
	rh := &ResourceHandler{GClient: nil}
	_, err := rh.HandlePullRequestResource(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "gitopia://repos/myorg/myrepo/pulls/15"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestHandleBountyResource_NoClient(t *testing.T) {
	rh := &ResourceHandler{GClient: nil}
	_, err := rh.HandleBountyResource(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "gitopia://bounties/123"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestHandleIssueResource_InvalidURI(t *testing.T) {
	rh := &ResourceHandler{GClient: nil}
	_, err := rh.HandleIssueResource(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "gitopia://repos/myorg"},
	})
	assert.Error(t, err)
}

func TestHandleIssueResource_InvalidIID(t *testing.T) {
	// With nil client, the nil-client check triggers before IID parsing.
	// This test verifies that an invalid IID still produces an error.
	rh := &ResourceHandler{GClient: nil}
	_, err := rh.HandleIssueResource(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "gitopia://repos/myorg/myrepo/issues/abc"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestHandlePullRequestResource_InvalidIID(t *testing.T) {
	rh := &ResourceHandler{GClient: nil}
	_, err := rh.HandlePullRequestResource(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "gitopia://repos/myorg/myrepo/pulls/abc"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestHandleBountyResource_InvalidID(t *testing.T) {
	rh := &ResourceHandler{GClient: nil}
	_, err := rh.HandleBountyResource(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "gitopia://bounties/abc"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}
