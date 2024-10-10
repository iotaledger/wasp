package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/suiconn"
)

type httpTransport struct {
	client *suiconn.HTTPClient
}

var _ transport = &httpTransport{}

func NewHTTP(url string) *Client {
	return &Client{
		transport: &httpTransport{
			client: suiconn.NewHTTPClient(url),
		},
	}
}

func (h *httpTransport) Call(ctx context.Context, v any, method suiconn.JsonRPCMethod, args ...any) error {
	return h.client.CallContext(ctx, v, method, args...)
}

func (h *httpTransport) Subscribe(
	ctx context.Context,
	v chan<- []byte,
	method suiconn.JsonRPCMethod,
	args ...any,
) error {
	panic("cannot subscribe over http")
}

func (h *httpTransport) WaitUntilStopped() {
}
