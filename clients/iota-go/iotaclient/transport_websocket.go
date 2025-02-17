package iotaclient

import (
	"context"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
)

type wsTransport struct {
	client *iotaconn.WebsocketClient
}

var _ Transport = &wsTransport{}

func NewWebsocket(
	ctx context.Context,
	wsURL string,
	waitUntilEffectsVisible *WaitParams,
	log *logger.Logger,
) (*Client, error) {
	ws, err := iotaconn.NewWebsocketClient(ctx, wsURL, log.Named("iotago-ws"))
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
