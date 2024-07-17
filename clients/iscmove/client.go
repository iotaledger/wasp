package iscmove

import (
	"context"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
)

type Config struct {
	APIURL   string
	GraphURL string
}

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	*suiclient.Client
	*SuiGraph

	config Config
}

func NewClient(config Config) *Client {
	return &Client{
		suiclient.New(config.APIURL),
		NewGraph(config.GraphURL),
		config,
	}
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	return suiclient.RequestFundsFromFaucet(ctx, address.AsSuiAddress(), suiconn.FaucetURL(c.config.APIURL))
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.GetLatestSuiSystemState(ctx)
	return err
}
