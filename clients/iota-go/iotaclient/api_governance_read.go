package iotaclient

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

func (c *Client) GetCommitteeInfo(
	ctx context.Context,
	epoch *iotajsonrpc.BigInt, // optional
) (*iotajsonrpc.CommitteeInfo, error) {
	var resp iotajsonrpc.CommitteeInfo
	if err := c.transport.Call(ctx, &resp, getCommitteeInfo, epoch); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error) {
	var resp iotajsonrpc.IotaSystemStateSummary
	if err := c.transport.Call(ctx, &resp, getLatestIotaSystemState); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error) {
	var resp iotajsonrpc.BigInt
	if err := c.transport.Call(ctx, &resp, getReferenceGasPrice); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error) {
	var resp []*iotajsonrpc.DelegatedStake
	if err := c.transport.Call(ctx, &resp, getStakes, owner); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetStakesByIds(ctx context.Context, stakedIotaIds []iotago.ObjectID) (
	[]*iotajsonrpc.DelegatedStake,
	error,
) {
	var resp []*iotajsonrpc.DelegatedStake
	if err := c.transport.Call(ctx, &resp, getStakesByIDs, stakedIotaIds); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error) {
	var resp iotajsonrpc.ValidatorsApy
	if err := c.transport.Call(ctx, &resp, getValidatorsApy); err != nil {
		return nil, err
	}
	return &resp, nil
}
