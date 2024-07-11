package iscmove

import (
	"context"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Config struct {
	APIURL   string
	GraphURL string
}

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
	var faucetURL string
	switch c.config.APIURL {
	case suiconn.TestnetEndpointURL:
		faucetURL = suiconn.TestnetFaucetURL
	case suiconn.DevnetEndpointURL:
		faucetURL = suiconn.DevnetFaucetURL
	case suiconn.LocalnetEndpointURL:
		faucetURL = suiconn.LocalnetFaucetURL
	default:
		panic("not supported network")
	}
	return suiclient.RequestFundsFromFaucet(context.TODO(), address.AsSuiAddress(), faucetURL)
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.GetLatestSuiSystemState(ctx)
	return err
}
