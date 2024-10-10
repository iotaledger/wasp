package iscmoveclient

import (
	"context"

	"github.com/iotaledger/hive.go/logger"
	iotaclient2 "github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	*iotaclient2.Client
	faucetURL string
}

func NewClient(client *iotaclient2.Client, faucetURL string) *Client {
	return &Client{
		Client:    client,
		faucetURL: faucetURL,
	}
}

func NewHTTPClient(apiURL, faucetURL string) *Client {
	return NewClient(
		iotaclient2.NewHTTP(apiURL),
		faucetURL,
	)
}

func NewWebsocketClient(
	ctx context.Context,
	wsURL, faucetURL string,
	log *logger.Logger,
) (*Client, error) {
	ws, err := iotaclient2.NewWebsocket(ctx, wsURL, log)
	if err != nil {
		return nil, err
	}
	return NewClient(ws, faucetURL), nil
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	if c.faucetURL == "" {
		panic("missing faucetURL")
	}
	return iotaclient2.RequestFundsFromFaucet(ctx, address.AsSuiAddress(), c.faucetURL)
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.GetLatestSuiSystemState(ctx)
	return err
}
