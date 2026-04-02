package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TrustLevel represents the permission tier for tool access.
type TrustLevel int

const (
	TrustReadOnly   TrustLevel = 1
	TrustLocalWrite TrustLevel = 2
	TrustChainWrite TrustLevel = 3
)

// ParseTrustLevel converts a string to a TrustLevel.
// Defaults to TrustChainWrite for backward compatibility.
func ParseTrustLevel(s string) TrustLevel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "readonly", "read_only", "ro":
		return TrustReadOnly
	case "localwrite", "local_write", "lw":
		return TrustLocalWrite
	case "chainwrite", "chain_write", "cw", "":
		return TrustChainWrite
	default:
		return TrustChainWrite
	}
}

// String returns the human-readable name of the trust level.
func (t TrustLevel) String() string {
	switch t {
	case TrustReadOnly:
		return "readonly"
	case TrustLocalWrite:
		return "localwrite"
	case TrustChainWrite:
		return "chainwrite"
	default:
		return "unknown"
	}
}

// toolTrustRequirements maps tool names to the minimum trust level required.
var toolTrustRequirements = map[string]TrustLevel{
	// Read-only tools
	"get_user_context":               TrustReadOnly,
	"set_active_dao":                 TrustReadOnly,
	"refresh_user_context":           TrustReadOnly,
	"claim_fee_grant":                TrustReadOnly,
	"create_user":            TrustChainWrite,
	"list_repos":             TrustReadOnly,
	"get_repo":               TrustReadOnly,
	"list_branches":          TrustReadOnly,
	"get_file_contents":      TrustReadOnly,
	"list_issues":            TrustReadOnly,
	"get_issue":              TrustReadOnly,
	"list_pull_requests":     TrustReadOnly,
	"get_pull_request":       TrustReadOnly,
	"get_pull_request_diff":  TrustReadOnly,
	"list_bounties":          TrustReadOnly,
	"get_bounty":             TrustReadOnly,
	"list_tags":              TrustReadOnly,
	"list_commits":           TrustReadOnly,
	"list_releases":          TrustReadOnly,
	"list_labels":            TrustReadOnly,
	"get_dao":                TrustReadOnly,
	"dao_list_members":       TrustReadOnly,
	"dao_list_proposals":     TrustReadOnly,
	"dao_get_proposal":       TrustReadOnly,

	// Local-write tools
	"git_clone":             TrustLocalWrite,
	"create_feature_branch": TrustLocalWrite,
	"sync_with_remote":      TrustLocalWrite,

	// Chain-write tools (default for anything not listed)
	"create_repo":                 TrustChainWrite,
	"create_issue":                TrustChainWrite,
	"create_pull_request":         TrustChainWrite,
	"merge_pull_request":          TrustChainWrite,
	"git_push":                    TrustChainWrite,
	"commit_and_push_changes":             TrustChainWrite,
	"create_feature_branch_pr":            TrustChainWrite,
	"update_feature_branch":               TrustChainWrite,
	"bootstrap_repo":                      TrustChainWrite,
	"create_release":              TrustChainWrite,
	"create_label":                TrustChainWrite,
	"delete_label":                TrustChainWrite,
	"create_dao":                  TrustChainWrite,
	"create_bounty":               TrustChainWrite,
	"update_bounty":               TrustChainWrite,
	"close_bounty":                TrustChainWrite,
	"delete_bounty":               TrustChainWrite,
	"comment_on_issue":            TrustChainWrite,
	"update_issue":                TrustChainWrite,
	"comment_on_pull_request":     TrustChainWrite,
	"fork_repository":             TrustChainWrite,
	"toggle_repository_forking":   TrustChainWrite,
	"dao_update_members":          TrustChainWrite,
	"dao_submit_proposal":         TrustChainWrite,
	"dao_vote":                    TrustChainWrite,
	"dao_exec":                    TrustChainWrite,
	"batch_execute":                       TrustChainWrite,
	"confirm_transaction":                 TrustChainWrite,
	"reject_transaction":                  TrustChainWrite,
	"list_pending_transactions":           TrustReadOnly,
}

// CheckTrust returns an error if the current trust level is insufficient for the tool.
func CheckTrust(currentLevel TrustLevel, toolName string) error {
	required, ok := toolTrustRequirements[toolName]
	if !ok {
		// Unknown tools default to chain-write (most restrictive)
		required = TrustChainWrite
	}
	if currentLevel < required {
		return fmt.Errorf("tool %q requires trust level %s, but current level is %s", toolName, required, currentLevel)
	}
	return nil
}

// withTrust wraps a tool handler with trust-level enforcement. If the
// configured trust level is insufficient for the named tool, the handler
// returns an ACCESS_DENIED error without executing the underlying function.
// Apply this at registration time so every tool call is gated.
func withTrust[In any](h *ToolHandler, toolName string, fn mcp.ToolHandlerFor[In, any]) mcp.ToolHandlerFor[In, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input In) (*mcp.CallToolResult, any, error) {
		if err := CheckTrust(h.TrustLvl, toolName); err != nil {
			return toolErrorf(ErrAuthFailed, "Access denied: %s", err)
		}
		return fn(ctx, req, input)
	}
}
