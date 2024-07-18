package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (c *Client) GetCommitteeInfo(
	ctx context.Context,
	epoch *suijsonrpc.BigInt, // optional

) (*suijsonrpc.CommitteeInfo, error) {
	var resp suijsonrpc.CommitteeInfo
	return &resp, c.transport.Call(ctx, &resp, getCommitteeInfo, epoch)
}

func (c *Client) GetLatestSuiSystemState(ctx context.Context) (*suijsonrpc.SuiSystemStateSummary, error) {
	var resp suijsonrpc.SuiSystemStateSummary
	return &resp, c.transport.Call(ctx, &resp, getLatestSuiSystemState)
}

func (c *Client) GetReferenceGasPrice(ctx context.Context) (*suijsonrpc.BigInt, error) {
	var resp suijsonrpc.BigInt
	return &resp, c.transport.Call(ctx, &resp, getReferenceGasPrice)
}

func (c *Client) GetStakes(ctx context.Context, owner *sui.Address) ([]*suijsonrpc.DelegatedStake, error) {
	var resp []*suijsonrpc.DelegatedStake
	return resp, c.transport.Call(ctx, &resp, getStakes, owner)
}

func (c *Client) GetStakesByIds(ctx context.Context, stakedSuiIds []sui.ObjectID) ([]*suijsonrpc.DelegatedStake, error) {
	var resp []*suijsonrpc.DelegatedStake
	return resp, c.transport.Call(ctx, &resp, getStakesByIds, stakedSuiIds)
}

func (c *Client) GetValidatorsApy(ctx context.Context) (*suijsonrpc.ValidatorsApy, error) {
	var resp suijsonrpc.ValidatorsApy
	return &resp, c.transport.Call(ctx, &resp, getValidatorsApy)
}