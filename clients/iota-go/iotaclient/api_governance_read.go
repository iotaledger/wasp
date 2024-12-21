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

func (c *Client) GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error) {
	var resp iotajsonrpc.IotaSystemStateSummary
	return &resp, c.transport.Call(ctx, &resp, getLatestIotaSystemState)
}

func (c *Client) GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error) {
	var resp iotajsonrpc.BigInt
	return &resp, c.transport.Call(ctx, &resp, getReferenceGasPrice)
}

func (c *Client) GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error) {
	var resp []*iotajsonrpc.DelegatedStake
	return resp, c.transport.Call(ctx, &resp, getStakes, owner)
}

func (c *Client) GetStakesByIds(ctx context.Context, stakedIotaIds []iotago.ObjectID) (
	[]*iotajsonrpc.DelegatedStake,
	error,
) {
	var resp []*iotajsonrpc.DelegatedStake
	return resp, c.transport.Call(ctx, &resp, getStakesByIDs, stakedIotaIds)
}

func (c *Client) GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error) {
	var resp iotajsonrpc.ValidatorsApy
	return &resp, c.transport.Call(ctx, &resp, getValidatorsApy)
}
