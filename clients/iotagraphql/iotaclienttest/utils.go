package iotaclienttest

import (
	"context"
	"errors"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

func GetValidatorAddress(ctx context.Context) (iotago.Address, error) {
	return iotago.Address{}, errors.New("not implemented")
}

func GetValidatorAddressWithCoins(ctx context.Context) (iotago.Address, error) {
	return iotago.Address{}, errors.New("not implemented")
}

func getCoinsPage(ctx context.Context, owner *iotago.Address, limit int, fetchCoinType *string) (*iotajsonrpc.CoinPage, error) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	if owner == nil {
		return nil, errors.New("owner address is required")
	}

	req := iotaclient.GetCoinsRequest{
		Owner:    owner,
		CoinType: fetchCoinType,
	}
	if limit > 0 {
		req.Limit = uint(limit)
	}

	return client.GetCoins(ctx, req)
}

func getCoins(ctx context.Context, owner *iotago.Address, limit int, fetchCoinType *string) (iotajsonrpc.Coins, error) {
	page, err := getCoinsPage(ctx, owner, limit, fetchCoinType)
	if err != nil {
		return nil, err
	}
	return iotajsonrpc.Coins(page.Data), nil
}
