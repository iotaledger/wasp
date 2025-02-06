package iotaclient

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
)

type wsTransport struct {
	client              *iotaconn.WebsocketClient
	url                 string
	log                 *logger.Logger
	ctx                 context.Context
	connectionRecreated chan struct{}
}

var _ transport = &wsTransport{}

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

	transport := &wsTransport{
		client:              ws,
		url:                 wsURL,
		log:                 log,
		ctx:                 ctx,
		connectionRecreated: make(chan struct{}),
	}

	go transport.handleReconnection()

	return &Client{
		transport:               transport,
		WaitUntilEffectsVisible: waitUntilEffectsVisible,
	}, nil
}

func (w *wsTransport) ConnectionRecreated() <-chan struct{} {
	return w.connectionRecreated
}

func (w *wsTransport) handleReconnection() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.client.Disconnected():
			w.log.Info("WebSocket connection lost, attempting to reconnect...")
			for {
				newClient, err := iotaconn.NewWebsocketClient(w.ctx, w.url, w.log)
				if err != nil {
					w.log.Warnf("Failed to reconnect: %v, retrying...", err)
					time.Sleep(time.Second * 1)
					continue
				}
				// oldClient := w.client
				w.client = newClient
				w.log.Info("Successfully reconnected WebSocket")
				// oldClient.WaitUntilStopped()
				w.connectionRecreated <- struct{}{}
				break
			}
		}
	}
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
