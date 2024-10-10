package iscmoveclient

import (
	"context"

	"github.com/iotaledger/hive.go/logger"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	*suiclient2.Client
	faucetURL string
}

func NewClient(client *suiclient2.Client, faucetURL string) *Client {
	return &Client{
		Client:    client,
		faucetURL: faucetURL,
	}
}

func NewHTTPClient(apiURL, faucetURL string) *Client {
	return NewClient(
		suiclient2.NewHTTP(apiURL),
		faucetURL,
	)
}

func NewWebsocketClient(
	ctx context.Context,
	wsURL, faucetURL string,
	log *logger.Logger,
) (*Client, error) {
	ws, err := suiclient2.NewWebsocket(ctx, wsURL, log)
	if err != nil {
		return nil, err
	}
	return NewClient(ws, faucetURL), nil
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	if c.faucetURL == "" {
		panic("missing faucetURL")
	}
	return suiclient2.RequestFundsFromFaucet(ctx, address.AsSuiAddress(), c.faucetURL)
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.GetLatestSuiSystemState(ctx)
	return err
}
