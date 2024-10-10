package iotaclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
)

type httpTransport struct {
	client *iotaconn.HTTPClient
}

var _ transport = &httpTransport{}

func NewHTTP(url string) *Client {
	return &Client{
		transport: &httpTransport{
			client: iotaconn.NewHTTPClient(url),
		},
	}
}

func (h *httpTransport) Call(ctx context.Context, v any, method iotaconn.JsonRPCMethod, args ...any) error {
	return h.client.CallContext(ctx, v, method, args...)
}

func (h *httpTransport) Subscribe(
	ctx context.Context,
	v chan<- []byte,
	method iotaconn.JsonRPCMethod,
	args ...any,
) error {
	panic("cannot subscribe over http")
}

func (h *httpTransport) WaitUntilStopped() {
}
