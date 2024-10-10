package iotaclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
)

type Client struct {
	transport transport
}

type transport interface {
	Call(ctx context.Context, v any, method iotaconn.JsonRPCMethod, args ...any) error
	Subscribe(ctx context.Context, v chan<- []byte, method iotaconn.JsonRPCMethod, args ...any) error
	WaitUntilStopped()
}

func (c *Client) WaitUntilStopped() {
	c.transport.WaitUntilStopped()
}
