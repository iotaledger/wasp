package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (s *Client) GetCommitteeInfo(
	ctx context.Context,
	epoch *suijsonrpc.BigInt, // optional

) (*suijsonrpc.CommitteeInfo, error) {
	var resp suijsonrpc.CommitteeInfo
	return &resp, s.http.CallContext(ctx, &resp, getCommitteeInfo, epoch)
}

func (s *Client) GetLatestSuiSystemState(ctx context.Context) (*suijsonrpc.SuiSystemStateSummary, error) {
	var resp suijsonrpc.SuiSystemStateSummary
	return &resp, s.http.CallContext(ctx, &resp, getLatestSuiSystemState)
}

func (s *Client) GetReferenceGasPrice(ctx context.Context) (*suijsonrpc.BigInt, error) {
	var resp suijsonrpc.BigInt
	return &resp, s.http.CallContext(ctx, &resp, getReferenceGasPrice)
}

func (s *Client) GetStakes(ctx context.Context, owner *sui.Address) ([]*suijsonrpc.DelegatedStake, error) {
	var resp []*suijsonrpc.DelegatedStake
	return resp, s.http.CallContext(ctx, &resp, getStakes, owner)
}

func (s *Client) GetStakesByIds(ctx context.Context, stakedSuiIds []sui.ObjectID) ([]*suijsonrpc.DelegatedStake, error) {
	var resp []*suijsonrpc.DelegatedStake
	return resp, s.http.CallContext(ctx, &resp, getStakesByIds, stakedSuiIds)
}

func (s *Client) GetValidatorsApy(ctx context.Context) (*suijsonrpc.ValidatorsApy, error) {
	var resp suijsonrpc.ValidatorsApy
	return &resp, s.http.CallContext(ctx, &resp, getValidatorsApy)
}