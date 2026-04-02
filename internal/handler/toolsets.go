package handler

import "strings"

// Toolset groups tools by functional area. Tools in the "core" set are
// always registered. Other sets can be enabled via TOOLSETS env var.
type Toolset string

const (
	ToolsetCore     Toolset = "core"
	ToolsetWorkflow Toolset = "workflow"
)

// toolsetMembership maps tool names to their toolset.
// Any tool NOT listed here defaults to ToolsetCore.
var toolsetMembership = map[string]Toolset{
	"bootstrap_repo":           ToolsetWorkflow,
	"create_feature_branch":    ToolsetWorkflow,
	"create_feature_branch_pr": ToolsetWorkflow,
	"update_feature_branch":    ToolsetWorkflow,
	"commit_and_push_changes":  ToolsetWorkflow,
}

// ParseToolsets parses a comma-separated toolset string.
// Empty string or "all" enables all toolsets.
func ParseToolsets(s string) map[Toolset]bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || s == "all" {
		return map[Toolset]bool{
			ToolsetCore:     true,
			ToolsetWorkflow: true,
		}
	}

	result := make(map[Toolset]bool)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		switch Toolset(part) {
		case ToolsetCore, ToolsetWorkflow:
			result[Toolset(part)] = true
		}
	}
	// Core is always enabled
	result[ToolsetCore] = true
	return result
}

// IsToolEnabled checks if a tool should be registered given the active toolsets.
func IsToolEnabled(toolName string, activeToolsets map[Toolset]bool) bool {
	ts, ok := toolsetMembership[toolName]
	if !ok {
		ts = ToolsetCore
	}
	return activeToolsets[ts]
}
