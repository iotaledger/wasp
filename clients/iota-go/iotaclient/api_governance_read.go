package iotaclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
)

func (c *Client) GetCommitteeInfo(
	ctx context.Context,
	epoch *iotajsonrpc.BigInt, // optional

) (*iotajsonrpc.CommitteeInfo, error) {
	var resp iotajsonrpc.CommitteeInfo
	return &resp, c.transport.Call(ctx, &resp, getCommitteeInfo, epoch)
}

func (c *Client) GetLatestSuiSystemState(ctx context.Context) (*iotajsonrpc.SuiSystemStateSummary, error) {
	var resp iotajsonrpc.SuiSystemStateSummary
	return &resp, c.transport.Call(ctx, &resp, getLatestSuiSystemState)
}

func (c *Client) GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error) {
	var resp iotajsonrpc.BigInt
	return &resp, c.transport.Call(ctx, &resp, getReferenceGasPrice)
}

func (c *Client) GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error) {
	var resp []*iotajsonrpc.DelegatedStake
	return resp, c.transport.Call(ctx, &resp, getStakes, owner)
}

func (c *Client) GetStakesByIds(ctx context.Context, stakedSuiIds []iotago.ObjectID) (
	[]*iotajsonrpc.DelegatedStake,
	error,
) {
	var resp []*iotajsonrpc.DelegatedStake
	return resp, c.transport.Call(ctx, &resp, getStakesByIds, stakedSuiIds)
}

func (c *Client) GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error) {
	var resp iotajsonrpc.ValidatorsApy
	return &resp, c.transport.Call(ctx, &resp, getValidatorsApy)
}
