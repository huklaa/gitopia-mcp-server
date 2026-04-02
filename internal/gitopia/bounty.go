package gitopia

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
)

// ListBounties returns up to `limit` bounties from the chain.
func (c *Client) ListBounties(ctx context.Context, limit uint64) ([]gitopiatypes.Bounty, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.BountyAll(ctx, &gitopiatypes.QueryAllBountyRequest{
		Pagination: &query.PageRequest{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	return resp.Bounty, nil
}

// GetBounty returns a single bounty by ID.
func (c *Client) GetBounty(ctx context.Context, bountyID uint64) (*gitopiatypes.Bounty, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.Bounty(ctx, &gitopiatypes.QueryGetBountyRequest{Id: bountyID})
	if err != nil {
		return nil, err
	}
	b := resp.Bounty
	return &b, nil
}

// CreateBounty signs & broadcasts a MsgCreateBounty.
func (c *Client) CreateBounty(
	ctx context.Context,
	w Wallet,
	msg *gitopiatypes.MsgCreateBounty,
) (*gitopiatypes.MsgCreateBountyResponse, error) {
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgCreateBountyResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateBounty signs & broadcasts a MsgUpdateBounty.
func (c *Client) UpdateBounty(
	ctx context.Context,
	w Wallet,
	msg *gitopiatypes.MsgUpdateBountyExpiry,
) (*gitopiatypes.MsgUpdateBountyExpiryResponse, error) {
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgUpdateBountyExpiryResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CloseBounty signs & broadcasts a MsgCloseBounty.
func (c *Client) CloseBounty(
	ctx context.Context,
	w Wallet,
	msg *gitopiatypes.MsgCloseBounty,
) (*gitopiatypes.MsgCloseBountyResponse, error) {
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgCloseBountyResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteBounty signs & broadcasts a MsgDeleteBounty.
func (c *Client) DeleteBounty(
	ctx context.Context,
	w Wallet,
	msg *gitopiatypes.MsgDeleteBounty,
) (*gitopiatypes.MsgDeleteBountyResponse, error) {
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgDeleteBountyResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
