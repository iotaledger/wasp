package iscmove

import (
	"context"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/suiclient"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	*suiclient.Client
	*SuiGraph
	faucetURL string
}

func NewClient(client *suiclient.Client, graphURL, faucetURL string) *Client {
	return &Client{
		Client:    client,
		SuiGraph:  NewGraph(graphURL),
		faucetURL: faucetURL,
	}
}

func NewHTTPClient(apiURL, graphURL, faucetURL string) *Client {
	return NewClient(
		suiclient.NewHTTP(apiURL),
		graphURL,
		faucetURL,
	)
}

func NewWebsocketClient(ctx context.Context, wsURL, graphURL, faucetURL string) *Client {
	return NewClient(
		suiclient.NewWebsocket(ctx, wsURL),
		graphURL,
		faucetURL,
	)
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	return suiclient.RequestFundsFromFaucet(ctx, address.AsSuiAddress(), c.faucetURL)
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.GetLatestSuiSystemState(ctx)
	return err
}
