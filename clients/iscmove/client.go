package iscmove

import (
	"context"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/suiclient"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	*suiclient.Client
	faucetURL string
}

func NewClient(client *suiclient.Client, faucetURL string) *Client {
	return &Client{
		Client:    client,
		faucetURL: faucetURL,
	}
}

func NewHTTPClient(apiURL, faucetURL string) *Client {
	return NewClient(
		suiclient.NewHTTP(apiURL),
		faucetURL,
	)
}

func NewWebsocketClient(
	ctx context.Context,
	wsURL, faucetURL string,
	log *logger.Logger,
) (*Client, error) {
	ws, err := suiclient.NewWebsocket(ctx, wsURL, log)
	if err != nil {
		return nil, err
	}
	return NewClient(ws, faucetURL), nil
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	if c.faucetURL == "" {
		panic("missing faucetURL")
	}
	return suiclient.RequestFundsFromFaucet(ctx, address.AsSuiAddress(), c.faucetURL)
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.GetLatestSuiSystemState(ctx)
	return err
}
