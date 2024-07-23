package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/suiconn"
)

type wsTransport struct {
	client *suiconn.WebsocketClient
}

var _ transport = &wsTransport{}

func NewWebsocket(ctx context.Context, wsURL string) *Client {
	return &Client{
		transport: &wsTransport{
			client: suiconn.NewWebsocketClient(ctx, wsURL),
		},
	}
}

func (w *wsTransport) Call(ctx context.Context, v any, method suiconn.JsonRPCMethod, args ...any) error {
	return w.client.CallContext(ctx, v, method, args...)
}

func (w *wsTransport) Subscribe(ctx context.Context, v chan<- []byte, method suiconn.JsonRPCMethod, args ...any) error {
	return w.client.Subscribe(ctx, v, method, args...)
}
