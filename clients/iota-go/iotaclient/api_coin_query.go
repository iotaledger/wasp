package iotaclient

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

func (c *Client) GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.Balance, error) {
	var resp []*iotajsonrpc.Balance
	if err := c.transport.Call(ctx, &resp, getAllBalances, owner); err != nil {
		return nil, err
	}
	return resp, nil
}

type GetAllCoinsRequest struct {
	Owner  *iotago.Address
	Cursor *iotago.ObjectID // optional
	Limit  uint             // optional
}

// start with the first object when cursor is nil
func (c *Client) GetAllCoins(ctx context.Context, req GetAllCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	var resp iotajsonrpc.CoinPage
	return &resp, c.transport.Call(ctx, &resp, getAllCoins, req.Owner, req.Cursor, req.Limit)
}

type GetBalanceRequest struct {
	Owner    *iotago.Address
	CoinType string // optional
}

// GetBalance to use default iotago coin(0x2::iota::IOTA) when coinType is empty
func (c *Client) GetBalance(ctx context.Context, req GetBalanceRequest) (*iotajsonrpc.Balance, error) {
	resp := iotajsonrpc.Balance{}
	if req.CoinType == "" {
		return &resp, c.transport.Call(ctx, &resp, getBalance, req.Owner)
	} else {
		return &resp, c.transport.Call(ctx, &resp, getBalance, req.Owner, req.CoinType)
	}
}

func (c *Client) GetCoinMetadata(ctx context.Context, coinType string) (*iotajsonrpc.IotaCoinMetadata, error) {
	var resp iotajsonrpc.IotaCoinMetadata
	return &resp, c.transport.Call(ctx, &resp, getCoinMetadata, coinType)
}

type GetCoinsRequest struct {
	Owner    *iotago.Address
	CoinType *string          // optional
	Cursor   *iotago.ObjectID // optional
	Limit    uint             // optional
}

// GetCoins to use default iotago coin(0x2::iota::IOTA) when coinType is nil
// start with the first object when cursor is nil
func (c *Client) GetCoins(ctx context.Context, req GetCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	var resp iotajsonrpc.CoinPage
	return &resp, c.transport.Call(ctx, &resp, getCoins, req.Owner, req.CoinType, req.Cursor, req.Limit)
}

func (c *Client) GetTotalSupply(ctx context.Context, coinType string) (*iotajsonrpc.Supply, error) {
	var resp iotajsonrpc.Supply
	return &resp, c.transport.Call(ctx, &resp, getTotalSupply, coinType)
}
