package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (s *Client) GetAllBalances(ctx context.Context, owner *sui.Address) ([]*suijsonrpc.Balance, error) {
	var resp []*suijsonrpc.Balance
	return resp, s.http.CallContext(ctx, &resp, getAllBalances, owner)
}

type GetAllCoinsRequest struct {
	Owner  *sui.Address
	Cursor *sui.ObjectID // optional
	Limit  uint          // optional
}

// start with the first object when cursor is nil
func (s *Client) GetAllCoins(ctx context.Context, req GetAllCoinsRequest) (*suijsonrpc.CoinPage, error) {
	var resp suijsonrpc.CoinPage
	return &resp, s.http.CallContext(ctx, &resp, getAllCoins, req.Owner, req.Cursor, req.Limit)
}

type GetBalanceRequest struct {
	Owner    *sui.Address
	CoinType sui.ObjectType // optional
}

// GetBalance to use default sui coin(0x2::sui::SUI) when coinType is empty
func (s *Client) GetBalance(ctx context.Context, req GetBalanceRequest) (*suijsonrpc.Balance, error) {
	resp := suijsonrpc.Balance{}
	if req.CoinType == "" {
		return &resp, s.http.CallContext(ctx, &resp, getBalance, req.Owner)
	} else {
		return &resp, s.http.CallContext(ctx, &resp, getBalance, req.Owner, req.CoinType)
	}
}

func (s *Client) GetCoinMetadata(ctx context.Context, coinType string) (*suijsonrpc.SuiCoinMetadata, error) {
	var resp suijsonrpc.SuiCoinMetadata
	return &resp, s.http.CallContext(ctx, &resp, getCoinMetadata, coinType)
}

type GetCoinsRequest struct {
	Owner    *sui.Address
	CoinType *sui.ObjectType // optional
	Cursor   *sui.ObjectID   // optional
	Limit    uint            // optional
}

// GetCoins to use default sui coin(0x2::sui::SUI) when coinType is nil
// start with the first object when cursor is nil
func (s *Client) GetCoins(ctx context.Context, req GetCoinsRequest) (*suijsonrpc.CoinPage, error) {
	var resp suijsonrpc.CoinPage
	return &resp, s.http.CallContext(ctx, &resp, getCoins, req.Owner, req.CoinType, req.Cursor, req.Limit)
}

func (s *Client) GetTotalSupply(ctx context.Context, coinType sui.ObjectType) (*suijsonrpc.Supply, error) {
	var resp suijsonrpc.Supply
	return &resp, s.http.CallContext(ctx, &resp, getTotalSupply, coinType)
}