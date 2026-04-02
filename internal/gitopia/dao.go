package gitopia

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	cosmosgroup "github.com/cosmos/cosmos-sdk/x/group"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
)

// CreateDao signs & broadcasts a MsgCreateDao.
func (c *Client) CreateDao(
	ctx context.Context,
	w Wallet,
	msg *gitopiatypes.MsgCreateDao,
) (*gitopiatypes.MsgCreateDaoResponse, error) {
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgCreateDaoResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetDao returns a DAO by name or address, including its group_id.
func (c *Client) GetDao(ctx context.Context, id string) (*gitopiatypes.Dao, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.Dao(ctx, &gitopiatypes.QueryGetDaoRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.Dao, nil
}

// ListGroupMembers returns members of a cosmos group by group ID.
func (c *Client) ListGroupMembers(ctx context.Context, groupID uint64, limit uint64) ([]*cosmosgroup.GroupMember, error) {
	q := cosmosgroup.NewQueryClient(c.conn)
	resp, err := q.GroupMembers(ctx, &cosmosgroup.QueryGroupMembersRequest{
		GroupId:    groupID,
		Pagination: &query.PageRequest{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	return resp.Members, nil
}

// ListProposalsByGroupPolicy returns proposals for a group policy address.
func (c *Client) ListProposalsByGroupPolicy(ctx context.Context, groupPolicyAddress string, limit uint64) ([]*cosmosgroup.Proposal, error) {
	q := cosmosgroup.NewQueryClient(c.conn)
	resp, err := q.ProposalsByGroupPolicy(ctx, &cosmosgroup.QueryProposalsByGroupPolicyRequest{
		Address:    groupPolicyAddress,
		Pagination: &query.PageRequest{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	return resp.Proposals, nil
}

// GetProposal returns a single group proposal by ID.
func (c *Client) GetProposal(ctx context.Context, proposalID uint64) (*cosmosgroup.Proposal, error) {
	q := cosmosgroup.NewQueryClient(c.conn)
	resp, err := q.Proposal(ctx, &cosmosgroup.QueryProposalRequest{
		ProposalId: proposalID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Proposal, nil
}

// GetGroupPoliciesByGroup returns group policies for a group ID.
func (c *Client) GetGroupPoliciesByGroup(ctx context.Context, groupID uint64) ([]*cosmosgroup.GroupPolicyInfo, error) {
	q := cosmosgroup.NewQueryClient(c.conn)
	resp, err := q.GroupPoliciesByGroup(ctx, &cosmosgroup.QueryGroupPoliciesByGroupRequest{
		GroupId:    groupID,
		Pagination: &query.PageRequest{Limit: 100},
	})
	if err != nil {
		return nil, err
	}
	return resp.GroupPolicies, nil
}

// UpdateGroupMembers signs & broadcasts a MsgUpdateGroupMembers.
func (c *Client) UpdateGroupMembers(ctx context.Context, w Wallet, admin string, groupID uint64, memberUpdates []cosmosgroup.MemberRequest) error {
	msg := &cosmosgroup.MsgUpdateGroupMembers{
		Admin:         admin,
		GroupId:       groupID,
		MemberUpdates: memberUpdates,
	}
	_, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	return err
}

// SubmitProposal signs & broadcasts a MsgSubmitProposal (text-only, no inner messages).
// Returns the new proposal ID.
func (c *Client) SubmitProposal(ctx context.Context, w Wallet, groupPolicyAddress string, proposers []string, metadata, title, summary string, execTry bool) (uint64, error) {
	exec := cosmosgroup.Exec_EXEC_UNSPECIFIED
	if execTry {
		exec = cosmosgroup.Exec_EXEC_TRY
	}
	msg := &cosmosgroup.MsgSubmitProposal{
		GroupPolicyAddress: groupPolicyAddress,
		Proposers:          proposers,
		Metadata:           metadata,
		Title:              title,
		Summary:            summary,
		Exec:               exec,
	}
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return 0, err
	}

	var resp cosmosgroup.MsgSubmitProposalResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return 0, err
	}
	return resp.ProposalId, nil
}

// VoteOnProposal signs & broadcasts a MsgVote.
func (c *Client) VoteOnProposal(ctx context.Context, w Wallet, proposalID uint64, option cosmosgroup.VoteOption, metadata string, execTry bool) error {
	exec := cosmosgroup.Exec_EXEC_UNSPECIFIED
	if execTry {
		exec = cosmosgroup.Exec_EXEC_TRY
	}
	msg := &cosmosgroup.MsgVote{
		ProposalId: proposalID,
		Voter:      w.Address(),
		Option:     option,
		Metadata:   metadata,
		Exec:       exec,
	}
	_, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	return err
}

// ExecProposal signs & broadcasts a MsgExec.
func (c *Client) ExecProposal(ctx context.Context, w Wallet, proposalID uint64) error {
	msg := &cosmosgroup.MsgExec{
		ProposalId: proposalID,
		Executor:   w.Address(),
	}
	_, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	return err
}
