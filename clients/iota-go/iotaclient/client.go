package iotaclient

import (
	"context"
	"time"

	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
)

type Client struct {
	transport transport

	// If WaitUntilEffectsVisible is set, it takes effect on any sent transaction with WaitForLocalExecution. It is
	// necessary because if the L1 node is overloaded, it may return an effects cert without actually having ececuted
	// the tx locally.
	WaitUntilEffectsVisible *WaitParams
}

type WaitParams struct {
	Attempts             int
	DelayBetweenAttempts time.Duration
}

var WaitForEffectsDisabled *WaitParams = nil
var WaitForEffectsEnabled *WaitParams = &WaitParams{
	Attempts:             5,
	DelayBetweenAttempts: 2 * time.Second,
}

type transport interface {
	Call(ctx context.Context, v any, method iotaconn.JsonRPCMethod, args ...any) error
	Subscribe(ctx context.Context, v chan<- []byte, method iotaconn.JsonRPCMethod, args ...any) error
	WaitUntilStopped()
}

func (c *Client) WaitUntilStopped() {
	c.transport.WaitUntilStopped()
}
