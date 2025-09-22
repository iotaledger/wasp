package iscmoveclient

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn_grpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
)

const eventBufferSize = 64

func pipeAndMapEvent[S any, D any](
	ctx context.Context,
	wg *sync.WaitGroup,
	in <-chan S,
	convert func(S) (D, bool),
) <-chan D {
	out := make(chan D, eventBufferSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case s, ok := <-in:
				if !ok {
					return
				}
				if d, ok := convert(s); ok {
					select {
					case out <- d:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}

type EventListener interface {
	SubscribeEvents(ctx context.Context) (<-chan iscmove.RequestEvent, error)
	SubscribeTransactions(ctx context.Context) (<-chan *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error)
	WaitUntilStopped()
}

type GRpcClientWrapper struct {
	eventClient       *iotaconn_grpc.StreamClient[*iotaconn_grpc.Event]
	transactionClient *iotaconn_grpc.StreamClient[*iotaconn_grpc.Transaction]
	httpClient        *Client
	wg                sync.WaitGroup
	log               log.Logger
}

func NewGRpcClientWrapper(log log.Logger, socketURL string, packageID iotago.PackageID, anchorID iotago.ObjectID, httpClient *Client) EventListener {
	return &GRpcClientWrapper{
		httpClient: httpClient,
		log:        log,
		eventClient: iotaconn_grpc.NewEventStreamClient(socketURL, &iotaconn_grpc.EventFilter{
			Filter: &iotaconn_grpc.EventFilter_MoveEventType{
				MoveEventType: &iotaconn_grpc.MoveEventTypeFilter{
					Module:    iscmove.RequestModuleName,
					Name:      iscmove.RequestEventObjectName,
					PackageId: &iotaconn_grpc.Address{Address: packageID.Bytes()},
				},
			},
		}, log),
		transactionClient: iotaconn_grpc.NewTransactionStreamClient(socketURL, &iotaconn_grpc.TransactionFilter{
			Filter: &iotaconn_grpc.TransactionFilter_ChangedObject{
				ChangedObject: &iotaconn_grpc.ChangedObjectFilter{
					ObjectId: &iotaconn_grpc.Address{Address: anchorID.Bytes()},
				},
			},
		}, log),
	}
}

func (g *GRpcClientWrapper) SubscribeEvents(ctx context.Context) (<-chan iscmove.RequestEvent, error) {
	raw := g.eventClient.Start(ctx)
	out := pipeAndMapEvent(ctx, &g.wg, raw, func(evt *iotaconn_grpc.Event) (iscmove.RequestEvent, bool) {
		req, err := bcs.Unmarshal[iscmove.RequestEvent](evt.GetEventData().GetData())
		if err != nil {
			g.log.LogErrorf("failed to unmarshal event: %v", err)
			return iscmove.RequestEvent{}, false
		}
		return req, true
	})
	return out, nil
}

func (g *GRpcClientWrapper) SubscribeTransactions(ctx context.Context) (<-chan *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error) {
	raw := g.transactionClient.Start(ctx)
	out := pipeAndMapEvent(ctx, &g.wg, raw, func(t *iotaconn_grpc.Transaction) (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], bool) {
		var effects serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]

		err := json.Unmarshal([]byte(t.EffectsJson), &effects)
		if err != nil {
			g.log.LogErrorf("failed to unmarshal transaction effects: %v", err)
			return nil, false
		}

		return &effects, true
	})
	return out, nil
}

func (g *GRpcClientWrapper) WaitUntilStopped() {
	g.wg.Wait()
}

type WebsocketClientWrapper struct {
	eventFilter       *iotajsonrpc.EventFilter
	transactionFilter *iotajsonrpc.TransactionFilter

	log   log.Logger
	wsURL string
	wg    sync.WaitGroup
}

func NewWebSocketClientWrapper(log log.Logger, wsURL string, packageID iotago.PackageID, anchorID iotago.ObjectID) EventListener {
	return &WebsocketClientWrapper{
		log:   log,
		wsURL: wsURL,
		eventFilter: &iotajsonrpc.EventFilter{
			And: &iotajsonrpc.AndOrEventFilter{
				Filter1: &iotajsonrpc.EventFilter{MoveEventType: &iotago.StructTag{
					Address: &packageID,
					Module:  iscmove.RequestModuleName,
					Name:    iscmove.RequestEventObjectName,
				}},
				Filter2: &iotajsonrpc.EventFilter{MoveEventField: &iotajsonrpc.EventFilterMoveEventField{
					Path:  iscmove.RequestEventAnchorFieldName,
					Value: anchorID.String(),
				}},
			},
		},
		transactionFilter: &iotajsonrpc.TransactionFilter{
			ChangedObject: &anchorID,
		},
	}
}

func startPush[T any](
	ctx context.Context,
	buf int,
	subscribe func(ctx context.Context, sink chan<- T) error,
) (<-chan T, error) {
	ch := make(chan T, buf)
	if err := subscribe(ctx, ch); err != nil {
		close(ch)
		return nil, err
	}
	return ch, nil
}

func (w *WebsocketClientWrapper) SubscribeEvents(ctx context.Context) (<-chan iscmove.RequestEvent, error) {
	wsClient, err := NewWebsocketClient(ctx, w.wsURL, "", iotaclient.WaitForEffectsEnabled, w.log)
	if err != nil {
		return nil, err
	}

	raw, err := startPush[*iotajsonrpc.IotaEvent](ctx, eventBufferSize, func(ctx context.Context, sink chan<- *iotajsonrpc.IotaEvent) error {
		return wsClient.SubscribeEvent(ctx, w.eventFilter, sink)
	})
	if err != nil {
		w.log.LogErrorf("failed to subscribe to events: %v", err)
		return nil, err
	}

	out := pipeAndMapEvent(ctx, &w.wg, raw, func(e *iotajsonrpc.IotaEvent) (iscmove.RequestEvent, bool) {
		req, err := bcs.Unmarshal[iscmove.RequestEvent](e.Bcs)
		if err != nil {
			w.log.LogErrorf("failed to unmarshal events: %v", err)
			return iscmove.RequestEvent{}, false
		}
		return req, true
	})
	return out, nil
}

func (w *WebsocketClientWrapper) SubscribeTransactions(ctx context.Context) (<-chan *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error) {
	wsClient, err := NewWebsocketClient(ctx, w.wsURL, "", iotaclient.WaitForEffectsEnabled, w.log)
	if err != nil {
		return nil, err
	}

	raw, err := startPush[*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]](
		ctx,
		eventBufferSize,
		func(ctx context.Context, sink chan<- *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]) error {
			return wsClient.SubscribeTransaction(ctx, w.transactionFilter, sink)
		},
	)
	if err != nil {
		w.log.LogErrorf("failed to subscribe to transactions: %v", err)
		return nil, err
	}

	out := pipeAndMapEvent(ctx, &w.wg, raw, func(e *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]) (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], bool) {
		return e, true
	})
	return out, nil
}

func (w *WebsocketClientWrapper) WaitUntilStopped() {
	w.wg.Wait()
}
