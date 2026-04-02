package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	cosmosgroup "github.com/cosmos/cosmos-sdk/x/group"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetDaoParams struct {
	ID string `json:"id" jsonschema:"DAO name or address"`
}

func (h *ToolHandler) GetDao(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params GetDaoParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	dao, err := h.GClient.GetDao(ctx, params.ID)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get DAO '%s': %s", params.ID, err)
	}

	out := map[string]any{
		"name":        dao.Name,
		"address":     dao.Address,
		"description": dao.Description,
		"avatar_url":  dao.AvatarUrl,
		"location":    dao.Location,
		"website":     dao.Website,
		"group_id":    dao.GroupId,
		"created_at":  dao.CreatedAt,
		"updated_at":  dao.UpdatedAt,
	}

	// Look up group policy address for convenience
	policies, err := h.GClient.GetGroupPoliciesByGroup(ctx, dao.GroupId)
	if err == nil && len(policies) > 0 {
		out["group_policy_address"] = policies[0].Address
	}

	return toolSuccessData(fmt.Sprintf("DAO '%s'", params.ID), out)
}

type CreateDaoParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Location    string `json:"location,omitempty"`
	Website     string `json:"website,omitempty"`
	Members     []struct {
		Address  string `json:"address"`
		Weight   string `json:"weight"`
		Metadata string `json:"metadata,omitempty"`
	} `json:"members,omitempty"`
	VotingPeriod string                  `json:"voting_period,omitempty"`
	Percentage   string                  `json:"percentage,omitempty"`
	Config       *gitopiatypes.DaoConfig `json:"config,omitempty"`
}

func (h *ToolHandler) CreateDao(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreateDaoParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("create_dao", p)
	}

	// Validate required parameters
	if err := validateCreateDaoParams(p); err != nil {
		return toolErrorf(ErrValidation, "Validation failed: %s", err)
	}

	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	// Set up members with defaults
	members := []cosmosgroup.MemberRequest{}

	// If no members provided, add creator as default member
	if len(p.Members) == 0 {
		members = append(members, cosmosgroup.MemberRequest{
			Address:  w.Address(),
			Weight:   "1", // Default weight of 1
			Metadata: "Creator and initial member",
		})
	} else {
		// Use provided members
		for _, m := range p.Members {
			members = append(members, cosmosgroup.MemberRequest{
				Address:  m.Address,
				Weight:   m.Weight,
				Metadata: m.Metadata,
			})
		}
	}

	// Set defaults for voting parameters
	votingPeriod := p.VotingPeriod
	if votingPeriod == "" {
		votingPeriod = "2" // 2 hours default
	}

	percentage := p.Percentage
	if percentage == "" {
		percentage = "0.50" // 50% default
	}

	msg := &gitopiatypes.MsgCreateDao{
		Creator:      w.Address(),
		Name:         p.Name,
		Description:  p.Description,
		AvatarUrl:    p.AvatarURL,
		Location:     p.Location,
		Website:      p.Website,
		Members:      members,
		VotingPeriod: votingPeriod,
		Percentage:   percentage,
		Config: &gitopiatypes.DaoConfig{
			RequirePullRequestProposal:        false,
			RequireRepositoryDeletionProposal: false,
			RequireCollaboratorProposal:       false,
			RequireReleaseProposal:            false,
		},
	}

	start := time.Now()
	res, err := h.GClient.CreateDao(ctx, w, msg)
	AuditLog("create_dao", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to create DAO: %s", err)
	}

	return toolSuccess("Created DAO", map[string]any{"dao_name": res.Id})
}

// validateCreateDaoParams validates the input parameters for DAO creation
func validateCreateDaoParams(params CreateDaoParams) error {
	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("DAO name cannot be empty")
	}
	if strings.TrimSpace(params.Description) == "" {
		return fmt.Errorf("DAO description cannot be empty")
	}

	// Validate members if provided
	for i, member := range params.Members {
		if strings.TrimSpace(member.Address) == "" {
			return fmt.Errorf("member %d address cannot be empty", i+1)
		}
		if strings.TrimSpace(member.Weight) == "" {
			return fmt.Errorf("member %d weight cannot be empty", i+1)
		}
	}

	return nil
}

// resolveDAOGroupID resolves a DAO name or address to its group_id.
// If groupID is already non-zero, returns it directly.
func (h *ToolHandler) resolveDAOGroupID(ctx context.Context, daoNameOrAddr string, groupID uint64) (uint64, error) {
	if groupID != 0 {
		return groupID, nil
	}
	if daoNameOrAddr == "" {
		return 0, fmt.Errorf("either 'dao' (name or address) or 'group_id' must be provided")
	}
	if h.GClient == nil {
		return 0, fmt.Errorf("failed to resolve DAO '%s': Gitopia client is not available", daoNameOrAddr)
	}
	dao, err := h.GClient.GetDao(ctx, daoNameOrAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve DAO '%s': %s", daoNameOrAddr, err)
	}
	return dao.GroupId, nil
}

// resolveDAOPolicyAddress resolves a DAO name or address to its group_policy_address.
// If policyAddress is already non-empty, returns it directly.
func (h *ToolHandler) resolveDAOPolicyAddress(ctx context.Context, daoNameOrAddr string, policyAddress string) (string, error) {
	if policyAddress != "" {
		return policyAddress, nil
	}
	if daoNameOrAddr == "" {
		return "", fmt.Errorf("either 'dao' (name or address) or 'group_policy_address' must be provided")
	}
	if h.GClient == nil {
		return "", fmt.Errorf("failed to resolve DAO '%s': Gitopia client is not available", daoNameOrAddr)
	}
	dao, err := h.GClient.GetDao(ctx, daoNameOrAddr)
	if err != nil {
		return "", fmt.Errorf("failed to resolve DAO '%s': %s", daoNameOrAddr, err)
	}
	policies, err := h.GClient.GetGroupPoliciesByGroup(ctx, dao.GroupId)
	if err != nil || len(policies) == 0 {
		return "", fmt.Errorf("DAO '%s' has no group policy address", daoNameOrAddr)
	}
	return policies[0].Address, nil
}

// ---- DAO Query Handlers ----

type ListDaoMembersParams struct {
	DAO     string `json:"dao,omitempty" jsonschema:"DAO name or address (alternative to group_id)"`
	GroupID uint64 `json:"group_id,omitempty" jsonschema:"The cosmos group ID of the DAO"`
	Limit   uint64 `json:"limit,omitempty" jsonschema:"Maximum number of members to return (default 100)"`
}

func (h *ToolHandler) ListDaoMembers(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ListDaoMembersParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	groupID, err := h.resolveDAOGroupID(ctx, p.DAO, p.GroupID)
	if err != nil {
		return toolErrorf(ErrValidation, "%s", err)
	}
	limit := p.Limit
	if limit == 0 {
		limit = 100
	}
	members, err := h.GClient.ListGroupMembers(ctx, groupID, limit)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list DAO members for group %d: %s", groupID, err)
	}
	out := make([]map[string]any, 0, len(members))
	for _, m := range members {
		entry := map[string]any{
			"group_id": m.GroupId,
		}
		if m.Member != nil {
			entry["address"] = m.Member.Address
			entry["weight"] = m.Member.Weight
			entry["metadata"] = m.Member.Metadata
			entry["added_at"] = m.Member.AddedAt.String()
		}
		out = append(out, entry)
	}
	return toolSuccessData(fmt.Sprintf("Found %d members for group %d", len(out), groupID), out)
}

type ListDaoProposalsParams struct {
	DAO                string `json:"dao,omitempty" jsonschema:"DAO name or address (alternative to group_policy_address)"`
	GroupPolicyAddress string `json:"group_policy_address,omitempty" jsonschema:"The group policy account address"`
	Limit              uint64 `json:"limit,omitempty" jsonschema:"Maximum number of proposals to return (default 50)"`
}

func (h *ToolHandler) ListDaoProposals(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ListDaoProposalsParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	policyAddr, err := h.resolveDAOPolicyAddress(ctx, p.DAO, p.GroupPolicyAddress)
	if err != nil {
		return toolErrorf(ErrValidation, "%s", err)
	}
	limit := p.Limit
	if limit == 0 {
		limit = 50
	}
	proposals, err := h.GClient.ListProposalsByGroupPolicy(ctx, policyAddr, limit)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list proposals for policy %s: %s", policyAddr, err)
	}
	out := make([]map[string]any, 0, len(proposals))
	for _, prop := range proposals {
		out = append(out, map[string]any{
			"id":                   prop.Id,
			"group_policy_address": prop.GroupPolicyAddress,
			"title":                prop.Title,
			"summary":              prop.Summary,
			"status":               prop.Status.String(),
			"proposers":            prop.Proposers,
			"submit_time":          prop.SubmitTime.String(),
			"voting_period_end":    prop.VotingPeriodEnd.String(),
		})
	}
	return toolSuccessData(fmt.Sprintf("Found %d proposals for policy %s", len(out), policyAddr), out)
}

type GetDaoProposalParams struct {
	ProposalID uint64 `json:"proposal_id" jsonschema:"The unique ID of the proposal"`
}

func (h *ToolHandler) GetDaoProposal(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p GetDaoProposalParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	proposal, err := h.GClient.GetProposal(ctx, p.ProposalID)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get proposal %d: %s", p.ProposalID, err)
	}
	out := map[string]any{
		"id":                   proposal.Id,
		"group_policy_address": proposal.GroupPolicyAddress,
		"metadata":             proposal.Metadata,
		"proposers":            proposal.Proposers,
		"submit_time":          proposal.SubmitTime.String(),
		"status":               proposal.Status.String(),
		"final_tally_result": map[string]string{
			"yes_count":          proposal.FinalTallyResult.YesCount,
			"abstain_count":      proposal.FinalTallyResult.AbstainCount,
			"no_count":           proposal.FinalTallyResult.NoCount,
			"no_with_veto_count": proposal.FinalTallyResult.NoWithVetoCount,
		},
		"voting_period_end": proposal.VotingPeriodEnd.String(),
		"executor_result":   proposal.ExecutorResult.String(),
		"title":             proposal.Title,
		"summary":           proposal.Summary,
	}
	return toolSuccessData(fmt.Sprintf("Proposal %d", p.ProposalID), out)
}

// ---- DAO Write Handlers ----

type UpdateDaoMembersParams struct {
	DAO           string `json:"dao,omitempty" jsonschema:"DAO name or address (alternative to group_id)"`
	GroupID       uint64 `json:"group_id,omitempty" jsonschema:"The cosmos group ID of the DAO"`
	MemberUpdates []struct {
		Address  string `json:"address"`
		Weight   string `json:"weight"`
		Metadata string `json:"metadata,omitempty"`
	} `json:"member_updates" jsonschema:"List of member updates. Set weight to 0 to remove a member."`
}

func (h *ToolHandler) UpdateDaoMembers(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p UpdateDaoMembersParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("dao_update_members", p)
	}

	// Resolve DAO name to group_id if needed
	groupID, err := h.resolveDAOGroupID(ctx, p.DAO, p.GroupID)
	if err != nil {
		return toolErrorf(ErrValidation, "%s", err)
	}
	if len(p.MemberUpdates) == 0 {
		return toolErrorf(ErrValidation, "at least one member update is required")
	}
	for i, m := range p.MemberUpdates {
		if strings.TrimSpace(m.Address) == "" {
			return toolErrorf(ErrValidation, "member_updates[%d] address cannot be empty", i)
		}
	}

	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
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

	msg := &cosmosgroup.MsgUpdateGroupMembers{
		Admin:         w.Address(),
		GroupId:       groupID,
		MemberUpdates: updates,
	}
	detail := fmt.Sprintf("group_id=%d", groupID)

	result, _, err := h.signOrHold(ctx, req, w, []sdk.Msg{msg}, "dao_update_members", detail)
	if result != nil || err != nil {
		return result, nil, err
	}

	return toolSuccess(fmt.Sprintf("Updated members for group %d", groupID), nil)
}

type SubmitDaoProposalParams struct {
	DAO                string `json:"dao,omitempty" jsonschema:"DAO name or address (alternative to group_policy_address)"`
	GroupPolicyAddress string `json:"group_policy_address,omitempty" jsonschema:"The group policy account address"`
	Title              string `json:"title" jsonschema:"Title of the proposal"`
	Summary            string `json:"summary,omitempty" jsonschema:"Summary of the proposal"`
	Metadata           string `json:"metadata,omitempty" jsonschema:"Arbitrary metadata for the proposal"`
	ExecTry            bool   `json:"exec_try,omitempty" jsonschema:"If true, try to execute immediately after submission"`
}

func (h *ToolHandler) SubmitDaoProposal(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p SubmitDaoProposalParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("dao_submit_proposal", p)
	}

	// Resolve DAO name to group_policy_address if needed
	policyAddr, err := h.resolveDAOPolicyAddress(ctx, p.DAO, p.GroupPolicyAddress)
	if err != nil {
		return toolErrorf(ErrValidation, "%s", err)
	}
	if strings.TrimSpace(p.Title) == "" {
		return toolErrorf(ErrValidation, "title is required")
	}

	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	exec := cosmosgroup.Exec_EXEC_UNSPECIFIED
	if p.ExecTry {
		exec = cosmosgroup.Exec_EXEC_TRY
	}
	msg := &cosmosgroup.MsgSubmitProposal{
		GroupPolicyAddress: policyAddr,
		Proposers:          []string{w.Address()},
		Metadata:           p.Metadata,
		Title:              p.Title,
		Summary:            p.Summary,
		Exec:               exec,
	}
	detail := fmt.Sprintf("policy=%s", policyAddr)

	result, txRespAny, err := h.signOrHold(ctx, req, w, []sdk.Msg{msg}, "dao_submit_proposal", detail)
	if result != nil || err != nil {
		return result, nil, err
	}

	// Parse proposal ID from broadcast response
	txResp, _ := txRespAny.(*tx.GetTxResponse)
	var resp cosmosgroup.MsgSubmitProposalResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResp, &resp); err != nil {
		return toolErrorf(ErrChainTx, "Proposal submitted but failed to parse response: %s", err)
	}

	return toolSuccess("Submitted proposal", map[string]any{"proposal_id": resp.ProposalId})
}

type VoteDaoProposalParams struct {
	ProposalID uint64 `json:"proposal_id" jsonschema:"The unique ID of the proposal to vote on"`
	Option     string `json:"option" jsonschema:"Vote option: yes, no, abstain, or no_with_veto"`
	Metadata   string `json:"metadata,omitempty" jsonschema:"Arbitrary metadata for the vote"`
	ExecTry    bool   `json:"exec_try,omitempty" jsonschema:"If true, try to execute proposal after voting"`
}

func (h *ToolHandler) VoteDaoProposal(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p VoteDaoProposalParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("dao_vote", p)
	}

	// Validate before wallet lookup
	voteOption, err := parseVoteOption(p.Option)
	if err != nil {
		return toolErrorf(ErrValidation, "%s", err)
	}

	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	exec := cosmosgroup.Exec_EXEC_UNSPECIFIED
	if p.ExecTry {
		exec = cosmosgroup.Exec_EXEC_TRY
	}
	msg := &cosmosgroup.MsgVote{
		ProposalId: p.ProposalID,
		Voter:      w.Address(),
		Option:     voteOption,
		Metadata:   p.Metadata,
		Exec:       exec,
	}
	detail := fmt.Sprintf("proposal=%d option=%s", p.ProposalID, p.Option)

	result, _, err := h.signOrHold(ctx, req, w, []sdk.Msg{msg}, "dao_vote", detail)
	if result != nil || err != nil {
		return result, nil, err
	}

	return toolSuccess(fmt.Sprintf("Voted '%s' on proposal %d", p.Option, p.ProposalID), nil)
}

type ExecDaoProposalParams struct {
	ProposalID uint64 `json:"proposal_id" jsonschema:"The unique ID of the proposal to execute"`
}

func (h *ToolHandler) ExecDaoProposal(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ExecDaoProposalParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("dao_exec", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	msg := &cosmosgroup.MsgExec{
		ProposalId: p.ProposalID,
		Executor:   w.Address(),
	}
	detail := fmt.Sprintf("proposal=%d", p.ProposalID)

	result, _, err := h.signOrHold(ctx, req, w, []sdk.Msg{msg}, "dao_exec", detail)
	if result != nil || err != nil {
		return result, nil, err
	}

	return toolSuccess(fmt.Sprintf("Executed proposal %d", p.ProposalID), nil)
}

// parseVoteOption converts a string to a cosmos group VoteOption.
func parseVoteOption(s string) (cosmosgroup.VoteOption, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "yes":
		return cosmosgroup.VOTE_OPTION_YES, nil
	case "no":
		return cosmosgroup.VOTE_OPTION_NO, nil
	case "abstain":
		return cosmosgroup.VOTE_OPTION_ABSTAIN, nil
	case "no_with_veto":
		return cosmosgroup.VOTE_OPTION_NO_WITH_VETO, nil
	default:
		return cosmosgroup.VOTE_OPTION_UNSPECIFIED, fmt.Errorf("invalid vote option %q: must be yes, no, abstain, or no_with_veto", s)
	}
}
