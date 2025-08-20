package iotaclient

import (
	"context"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
)

type wsTransport struct {
	client *iotaconn.WebsocketClient
}

var _ transport = &wsTransport{}

func NewWebsocket(
	ctx context.Context,
	wsURL string,
	waitUntilEffectsVisible *WaitParams,
	log log.Logger,
) (*Client, error) {
	ws, err := iotaconn.NewWebsocketClient(ctx, wsURL, log.NewChildLogger("iotago-ws"))
	if err != nil {
		return nil, err
	}
	return &Client{
		transport:               &wsTransport{client: ws},
		WaitUntilEffectsVisible: waitUntilEffectsVisible,
	}, nil
}

func (w *wsTransport) Call(ctx context.Context, v any, method iotaconn.JsonRPCMethod, args ...any) error {
	return w.client.CallContext(ctx, v, method, args...)
}

func (w *wsTransport) Subscribe(
	ctx context.Context,
	v chan<- []byte,
	method iotaconn.JsonRPCMethod,
	args ...any,
) error {
	return w.client.Subscribe(ctx, v, method, args...)
}

func (w *wsTransport) WaitUntilStopped() {
	w.client.WaitUntilStopped()
}
