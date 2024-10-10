package suiclient

import (
	"context"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/clients/iota-go/suiconn"
)

type wsTransport struct {
	client *suiconn.WebsocketClient
}

var _ transport = &wsTransport{}

func NewWebsocket(
	ctx context.Context,
	wsURL string,
	log *logger.Logger,
) (*Client, error) {
	ws, err := suiconn.NewWebsocketClient(ctx, wsURL, log.Named("sui-ws"))
	if err != nil {
		return nil, err
	}
	return &Client{transport: &wsTransport{client: ws}}, nil
}

func (w *wsTransport) Call(ctx context.Context, v any, method suiconn.JsonRPCMethod, args ...any) error {
	return w.client.CallContext(ctx, v, method, args...)
}

func (w *wsTransport) Subscribe(ctx context.Context, v chan<- []byte, method suiconn.JsonRPCMethod, args ...any) error {
	return w.client.Subscribe(ctx, v, method, args...)
}

func (w *wsTransport) WaitUntilStopped() {
	w.client.WaitUntilStopped()
}
