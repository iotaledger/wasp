package iotaclient

import (
	"context"
	"time"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

type CoinReader interface {
	GetCoins(ctx context.Context, req GetCoinsRequest) (*iotajsonrpc.CoinPage, error)
}

func WaitForCoins(
	ctx context.Context,
	reader CoinReader,
	owner *iotago.Address,
	limit int,
	timeout time.Duration,
) (*iotajsonrpc.CoinPage, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		cp, err := reader.GetCoins(ctx, GetCoinsRequest{Owner: owner, Limit: limit})
		if err == nil && len(cp.Data) > 0 {
			return cp, nil
		}
		if err != nil {
			lastErr = err
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return reader.GetCoins(ctx, GetCoinsRequest{Owner: owner, Limit: limit})
}
