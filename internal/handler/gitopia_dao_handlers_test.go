package handler

import (
	"context"
	"testing"

	cosmosgroup "github.com/cosmos/cosmos-sdk/x/group"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDao_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.CreateDao(context.Background(), &mcp.CallToolRequest{}, CreateDaoParams{
		Name: "test-dao", Description: "A test DAO",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestCreateDao_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.CreateDao(context.Background(), &mcp.CallToolRequest{}, CreateDaoParams{
		Name: "test-dao", Description: "A test DAO",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestCreateDao_ValidationEmptyName(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.CreateDao(context.Background(), &mcp.CallToolRequest{}, CreateDaoParams{
		Name: "", Description: "A test DAO",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "name cannot be empty")
}

func TestValidateCreateDaoParams(t *testing.T) {
	testCases := []struct {
		name        string
		params      CreateDaoParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid params",
			params: CreateDaoParams{
				Name:        "Test DAO",
				Description: "A test DAO description.",
				Members: []struct {
					Address  string `json:"address"`
					Weight   string `json:"weight"`
					Metadata string `json:"metadata,omitempty"`
				}{
					{Address: "gitopia1address1", Weight: "1"},
				},
			},
			expectError: false,
		},
		{
			name: "empty dao name",
			params: CreateDaoParams{
				Name:        "",
				Description: "A test DAO description.",
			},
			expectError: true,
			errorMsg:    "DAO name cannot be empty",
		},
		{
			name: "empty dao description",
			params: CreateDaoParams{
				Name:        "Test DAO",
				Description: "",
			},
			expectError: true,
			errorMsg:    "DAO description cannot be empty",
		},
		{
			name: "empty member address",
			params: CreateDaoParams{
				Name:        "Test DAO",
				Description: "A test DAO description.",
				Members: []struct {
					Address  string `json:"address"`
					Weight   string `json:"weight"`
					Metadata string `json:"metadata,omitempty"`
				}{
					{Address: "", Weight: "1"},
				},
			},
			expectError: true,
			errorMsg:    "member 1 address cannot be empty",
		},
		{
			name: "empty member weight",
			params: CreateDaoParams{
				Name:        "Test DAO",
				Description: "A test DAO description.",
				Members: []struct {
					Address  string `json:"address"`
					Weight   string `json:"weight"`
					Metadata string `json:"metadata,omitempty"`
				}{
					{Address: "gitopia1address1", Weight: ""},
				},
			},
			expectError: true,
			errorMsg:    "member 1 weight cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCreateDaoParams(tc.params)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- DAO Name Resolution Tests ----

func TestResolveDAOGroupID_DirectGroupID(t *testing.T) {
	h := &ToolHandler{}
	id, err := h.resolveDAOGroupID(context.Background(), "", 42)
	require.NoError(t, err)
	assert.Equal(t, uint64(42), id)
}

func TestResolveDAOGroupID_BothEmpty(t *testing.T) {
	h := &ToolHandler{}
	_, err := h.resolveDAOGroupID(context.Background(), "", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either 'dao' (name or address) or 'group_id' must be provided")
}

func TestResolveDAOGroupID_NameWithoutClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	_, err := h.resolveDAOGroupID(context.Background(), "my-dao", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve DAO")
}

func TestResolveDAOPolicyAddress_DirectAddress(t *testing.T) {
	h := &ToolHandler{}
	addr, err := h.resolveDAOPolicyAddress(context.Background(), "", "gitopia1abc123")
	require.NoError(t, err)
	assert.Equal(t, "gitopia1abc123", addr)
}

func TestResolveDAOPolicyAddress_BothEmpty(t *testing.T) {
	h := &ToolHandler{}
	_, err := h.resolveDAOPolicyAddress(context.Background(), "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either 'dao' (name or address) or 'group_policy_address' must be provided")
}

func TestListDaoMembers_WithDAOParam_NoClient(t *testing.T) {
	// When GClient is nil, the handler returns "not available" before resolving DAO
	h := &ToolHandler{GClient: nil}
	result, _, err := h.ListDaoMembers(context.Background(), &mcp.CallToolRequest{}, ListDaoMembersParams{DAO: "my-dao"})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestSubmitDaoProposal_WithDAOParam_NoClient(t *testing.T) {
	// SubmitDaoProposal hits rate limiter / dry-run / resolver before wallet check
	// With nil client, the resolver returns an error
	h := &ToolHandler{GClient: nil, DryRunMode: false}
	result, _, err := h.SubmitDaoProposal(context.Background(), &mcp.CallToolRequest{}, SubmitDaoProposalParams{DAO: "my-dao", Title: "test"})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- DAO Query Handler Tests ----

func TestListDaoMembers_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.ListDaoMembers(context.Background(), &mcp.CallToolRequest{}, ListDaoMembersParams{GroupID: 1})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestListDaoProposals_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.ListDaoProposals(context.Background(), &mcp.CallToolRequest{}, ListDaoProposalsParams{GroupPolicyAddress: "addr1"})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

func TestGetDaoProposal_NoClient(t *testing.T) {
	h := &ToolHandler{GClient: nil}
	result, _, err := h.GetDaoProposal(context.Background(), &mcp.CallToolRequest{}, GetDaoProposalParams{ProposalID: 1})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "not available")
}

// ---- DAO Write Handler Tests ----

func TestUpdateDaoMembers_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.UpdateDaoMembers(context.Background(), &mcp.CallToolRequest{}, UpdateDaoMembersParams{
		GroupID: 1,
		MemberUpdates: []struct {
			Address  string `json:"address"`
			Weight   string `json:"weight"`
			Metadata string `json:"metadata,omitempty"`
		}{{Address: "addr1", Weight: "1"}},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestUpdateDaoMembers_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.UpdateDaoMembers(context.Background(), &mcp.CallToolRequest{}, UpdateDaoMembersParams{
		GroupID: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestUpdateDaoMembers_ValidationEmptyGroupID(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.UpdateDaoMembers(context.Background(), &mcp.CallToolRequest{}, UpdateDaoMembersParams{
		GroupID: 0,
		MemberUpdates: []struct {
			Address  string `json:"address"`
			Weight   string `json:"weight"`
			Metadata string `json:"metadata,omitempty"`
		}{{Address: "addr1", Weight: "1"}},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "either 'dao' (name or address) or 'group_id' must be provided")
}

func TestUpdateDaoMembers_ValidationEmptyUpdates(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.UpdateDaoMembers(context.Background(), &mcp.CallToolRequest{}, UpdateDaoMembersParams{
		GroupID:       1,
		MemberUpdates: nil,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "at least one member update")
}

func TestUpdateDaoMembers_ValidationEmptyAddress(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.UpdateDaoMembers(context.Background(), &mcp.CallToolRequest{}, UpdateDaoMembersParams{
		GroupID: 1,
		MemberUpdates: []struct {
			Address  string `json:"address"`
			Weight   string `json:"weight"`
			Metadata string `json:"metadata,omitempty"`
		}{{Address: "", Weight: "1"}},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "address cannot be empty")
}

func TestSubmitDaoProposal_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.SubmitDaoProposal(context.Background(), &mcp.CallToolRequest{}, SubmitDaoProposalParams{
		GroupPolicyAddress: "policy1", Title: "Test Proposal",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestSubmitDaoProposal_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.SubmitDaoProposal(context.Background(), &mcp.CallToolRequest{}, SubmitDaoProposalParams{
		GroupPolicyAddress: "policy1", Title: "Test Proposal",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSubmitDaoProposal_ValidationEmptyTitle(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.SubmitDaoProposal(context.Background(), &mcp.CallToolRequest{}, SubmitDaoProposalParams{
		GroupPolicyAddress: "policy1", Title: "",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "title is required")
}

func TestSubmitDaoProposal_ValidationEmptyAddress(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.SubmitDaoProposal(context.Background(), &mcp.CallToolRequest{}, SubmitDaoProposalParams{
		GroupPolicyAddress: "", Title: "Test",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "either 'dao' (name or address) or 'group_policy_address' must be provided")
}

func TestVoteDaoProposal_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.VoteDaoProposal(context.Background(), &mcp.CallToolRequest{}, VoteDaoProposalParams{
		ProposalID: 1, Option: "yes",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestVoteDaoProposal_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.VoteDaoProposal(context.Background(), &mcp.CallToolRequest{}, VoteDaoProposalParams{
		ProposalID: 1, Option: "yes",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestVoteDaoProposal_InvalidOption(t *testing.T) {
	h := &ToolHandler{}
	result, _, err := h.VoteDaoProposal(context.Background(), &mcp.CallToolRequest{}, VoteDaoProposalParams{
		ProposalID: 1, Option: "invalid",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "invalid vote option")
}

func TestExecDaoProposal_DryRun(t *testing.T) {
	h := &ToolHandler{DryRunMode: true}
	result, _, err := h.ExecDaoProposal(context.Background(), &mcp.CallToolRequest{}, ExecDaoProposalParams{
		ProposalID: 1,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "dry_run")
}

func TestExecDaoProposal_RateLimited(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	h := &ToolHandler{ChainLimiter: limiter}
	result, _, err := h.ExecDaoProposal(context.Background(), &mcp.CallToolRequest{}, ExecDaoProposalParams{
		ProposalID: 1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ---- parseVoteOption Tests ----

func TestParseVoteOption(t *testing.T) {
	tests := []struct {
		input    string
		expected cosmosgroup.VoteOption
		wantErr  bool
	}{
		{"yes", cosmosgroup.VOTE_OPTION_YES, false},
		{"YES", cosmosgroup.VOTE_OPTION_YES, false},
		{"no", cosmosgroup.VOTE_OPTION_NO, false},
		{"abstain", cosmosgroup.VOTE_OPTION_ABSTAIN, false},
		{"no_with_veto", cosmosgroup.VOTE_OPTION_NO_WITH_VETO, false},
		{"invalid", cosmosgroup.VOTE_OPTION_UNSPECIFIED, true},
		{"", cosmosgroup.VOTE_OPTION_UNSPECIFIED, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseVoteOption(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
