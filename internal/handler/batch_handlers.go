package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosgroup "github.com/cosmos/cosmos-sdk/x/group"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BatchOperation represents a single operation in a batch transaction.
type BatchOperation struct {
	Tool   string         `json:"tool" jsonschema:"The tool name to execute"`
	Params map[string]any `json:"params" jsonschema:"Parameters for the tool as a JSON object"`
}

// BatchExecuteParams holds parameters for the batch_execute tool.
type BatchExecuteParams struct {
	Operations []BatchOperation `json:"operations" jsonschema:"List of operations to execute in a single transaction (max 10)"`
}

// BatchExecute executes multiple operations in a single on-chain transaction.
func (h *ToolHandler) BatchExecute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p BatchExecuteParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("batch_execute", p)
	}

	// Validate before wallet lookup
	if len(p.Operations) == 0 {
		return toolErrorf(ErrValidation, "at least one operation is required")
	}
	if len(p.Operations) > 10 {
		return toolErrorf(ErrValidation, "maximum 10 operations per batch")
	}

	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	msgs, err := buildBatchMessages(ctx, w, p.Operations)
	if err != nil {
		return toolErrorf(ErrValidation, "Failed to build batch messages: %s", err)
	}

	detail := fmt.Sprintf("ops=%d", len(p.Operations))
	result, _, err := h.signOrHold(ctx, req, w, msgs, "batch_execute", detail)
	if result != nil || err != nil {
		return result, nil, err
	}

	return toolSuccess(fmt.Sprintf("Executed batch transaction with %d operations", len(p.Operations)), map[string]any{
		"operations_count": len(p.Operations),
	})
}

// buildBatchMessages converts a list of batch operations into SDK messages.
func buildBatchMessages(ctx context.Context, w signing.Wallet, ops []BatchOperation) ([]sdk.Msg, error) {
	var msgs []sdk.Msg
	for i, op := range ops {
		opMsgs, err := buildSingleMessage(ctx, w, op)
		if err != nil {
			return nil, fmt.Errorf("operation[%d] (%s): %w", i, op.Tool, err)
		}
		msgs = append(msgs, opMsgs...)
	}
	return msgs, nil
}

// buildSingleMessage constructs SDK messages for a single batch operation.
// Only a curated set of tools is supported for security.
func buildSingleMessage(_ context.Context, w signing.Wallet, op BatchOperation) ([]sdk.Msg, error) {
	// Marshal params back to JSON for typed unmarshaling
	paramsJSON, err := json.Marshal(op.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	switch op.Tool {
	case "dao_vote":
		var p struct {
			ProposalID uint64 `json:"proposal_id"`
			Option     string `json:"option"`
			Metadata   string `json:"metadata"`
			ExecTry    bool   `json:"exec_try"`
		}
		if err := json.Unmarshal(paramsJSON, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		voteOption, err := parseVoteOption(p.Option)
		if err != nil {
			return nil, err
		}
		exec := cosmosgroup.Exec_EXEC_UNSPECIFIED
		if p.ExecTry {
			exec = cosmosgroup.Exec_EXEC_TRY
		}
		return []sdk.Msg{&cosmosgroup.MsgVote{
			ProposalId: p.ProposalID,
			Voter:      w.Address(),
			Option:     voteOption,
			Metadata:   p.Metadata,
			Exec:       exec,
		}}, nil

	case "dao_update_members":
		var p struct {
			GroupID       uint64 `json:"group_id"`
			MemberUpdates []struct {
				Address  string `json:"address"`
				Weight   string `json:"weight"`
				Metadata string `json:"metadata"`
			} `json:"member_updates"`
		}
		if err := json.Unmarshal(paramsJSON, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		updates := make([]cosmosgroup.MemberRequest, len(p.MemberUpdates))
		for i, m := range p.MemberUpdates {
			weight := m.Weight
			if weight == "" {
				weight = "1"
			}
			updates[i] = cosmosgroup.MemberRequest{
				Address:  m.Address,
				Weight:   weight,
				Metadata: m.Metadata,
			}
		}
		return []sdk.Msg{&cosmosgroup.MsgUpdateGroupMembers{
			Admin:         w.Address(),
			GroupId:       p.GroupID,
			MemberUpdates: updates,
		}}, nil

	case "comment_on_issue":
		var p struct {
			RepoID   uint64 `json:"repo_id"`
			IssueIID uint64 `json:"issue_iid"`
			Body     string `json:"body"`
		}
		if err := json.Unmarshal(paramsJSON, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return []sdk.Msg{gitopiatypes.NewMsgCreateComment(
			w.Address(), p.RepoID, p.IssueIID,
			gitopiatypes.CommentParentIssue,
			p.Body, nil, "", "", 0,
		)}, nil

	case "comment_on_pull_request":
		var p struct {
			RepoID  uint64 `json:"repo_id"`
			PullIID uint64 `json:"pull_iid"`
			Body    string `json:"body"`
		}
		if err := json.Unmarshal(paramsJSON, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return []sdk.Msg{gitopiatypes.NewMsgCreateComment(
			w.Address(), p.RepoID, p.PullIID,
			gitopiatypes.CommentParentPullRequest,
			p.Body, nil, "", "", 0,
		)}, nil

	default:
		return nil, fmt.Errorf("tool %q is not supported in batch operations; supported: dao_vote, dao_update_members, comment_on_issue, comment_on_pull_request",
			strings.TrimSpace(op.Tool))
	}
}
