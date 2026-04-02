package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToolsets_Empty(t *testing.T) {
	ts := ParseToolsets("")
	assert.True(t, ts[ToolsetCore])
	assert.True(t, ts[ToolsetWorkflow])
}

func TestParseToolsets_All(t *testing.T) {
	ts := ParseToolsets("all")
	assert.True(t, ts[ToolsetCore])
	assert.True(t, ts[ToolsetWorkflow])
}

func TestParseToolsets_CoreOnly(t *testing.T) {
	ts := ParseToolsets("core")
	assert.True(t, ts[ToolsetCore])
	assert.False(t, ts[ToolsetWorkflow])
}

func TestParseToolsets_CoreAndWorkflow(t *testing.T) {
	ts := ParseToolsets("core,workflow")
	assert.True(t, ts[ToolsetCore])
	assert.True(t, ts[ToolsetWorkflow])
}

func TestParseToolsets_InvalidIgnored(t *testing.T) {
	ts := ParseToolsets("core,nonexistent")
	assert.True(t, ts[ToolsetCore])
	assert.False(t, ts[ToolsetWorkflow])
}

func TestIsToolEnabled_CoreTool(t *testing.T) {
	ts := ParseToolsets("core")
	assert.True(t, IsToolEnabled("create_issue", ts))
	assert.True(t, IsToolEnabled("list_repos", ts))
	assert.True(t, IsToolEnabled("get_user_context", ts))
}

func TestIsToolEnabled_WorkflowToolBlocked(t *testing.T) {
	ts := ParseToolsets("core")
	assert.False(t, IsToolEnabled("bootstrap_repo", ts))
	assert.False(t, IsToolEnabled("create_feature_branch_pr", ts))
	assert.False(t, IsToolEnabled("commit_and_push_changes", ts))
}

func TestIsToolEnabled_WorkflowToolAllowed(t *testing.T) {
	ts := ParseToolsets("core,workflow")
	assert.True(t, IsToolEnabled("bootstrap_repo", ts))
	assert.True(t, IsToolEnabled("create_feature_branch_pr", ts))
}
