package iscmoveclient

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn_grpc"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/transaction"
)

type EventClient interface {
	SubscribeEvents(ctx context.Context) (<-chan iscmove.RequestEvent, error)
	WaitUntilStopped()
}

type GRpcClientWrapper struct {
	client   *iotaconn_grpc.EventStreamClient
	anchorID iotago.ObjectID

	wg sync.WaitGroup
}

const eventBufferSize = 64

func (g *GRpcClientWrapper) SubscribeEvents(ctx context.Context) (<-chan iscmove.RequestEvent, error) {
	iotaEvents := g.client.Start(ctx)
	out := make(chan iscmove.RequestEvent, eventBufferSize)

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-iotaEvents:
				if !ok {
					return
				}
				fmt.Println(hexutil.Encode(evt.EventData.Data))
				reqEvent, err := bcs.Unmarshal[iscmove.RequestEvent](evt.EventData.Data)
				if err != nil {
					continue
				}
				select {
				case out <- reqEvent:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

func (g *GRpcClientWrapper) WaitUntilStopped() {
	g.wg.Wait()
}

func NewGRpcClientWrapper(log log.Logger, socketURL string, packageID iotago.PackageID, anchorID iotago.ObjectID) EventClient {
	return &GRpcClientWrapper{
		client: iotaconn_grpc.NewEventStreamClient(socketURL, &iotaconn_grpc.EventFilter{
			Filter: &iotaconn_grpc.EventFilter_MoveEventType{
				MoveEventType: &iotaconn_grpc.MoveEventTypeFilter{
					Module:    iscmove.RequestModuleName,
					Name:      iscmove.RequestEventObjectName,
					PackageId: &iotaconn_grpc.Address{Address: packageID.Bytes()},
				},
			},
		}, log),
		anchorID: anchorID,
	}
}

type WebsocketClientWrapper struct {
	client *Client
	filter *iotajsonrpc.EventFilter
	log    log.Logger
	wsURL  string
	wg     sync.WaitGroup
}

func NewWebSocketClientWrapper(log log.Logger, wsURL string, packageID iotago.PackageID, anchorID iotago.ObjectID) EventClient {
	return &WebsocketClientWrapper{
		log:   log,
		wsURL: wsURL,
		filter: &iotajsonrpc.EventFilter{
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
	}
}

func (w *WebsocketClientWrapper) SubscribeEvents(ctx context.Context) (<-chan iscmove.RequestEvent, error) {
	ev := make(chan *iotajsonrpc.IotaEvent, eventBufferSize)

	client, err := NewWebsocketClient(ctx, w.wsURL, "", iotaclient.WaitForEffectsEnabled, w.log)
	if err != nil {
		return nil, err
	}

	w.client = client

	if err := w.client.SubscribeEvent(ctx, w.filter, ev); err != nil {
		return nil, err
	}

	out := make(chan iscmove.RequestEvent, eventBufferSize)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case iotaEvent, ok := <-ev:
				if !ok {
					return
				}
				reqEvent, err := bcs.Unmarshal[iscmove.RequestEvent](iotaEvent.Bcs)
				if err != nil {
					continue
				}

				select {
				case out <- reqEvent:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

func (w *WebsocketClientWrapper) WaitUntilStopped() {
	w.wg.Wait()
}

type ChainFeed struct {
	eventClient   EventClient
	httpClient    *Client
	iscPackageID  iotago.PackageID
	anchorAddress iotago.ObjectID
	log           log.Logger
}

func selectEventClient(log log.Logger, socketURL string, iscPackageID iotago.PackageID, anchorID iotago.ObjectID) (EventClient, error) {
	if strings.HasPrefix(socketURL, "grpc://") {
		return NewGRpcClientWrapper(log, strings.ReplaceAll(socketURL, "grpc://", ""), iscPackageID, anchorID), nil
	} else if strings.HasPrefix(socketURL, "ws://") || strings.HasPrefix(socketURL, "wss://") {
		return NewWebSocketClientWrapper(log, socketURL, iscPackageID, anchorID), nil
	} else {
		return nil, fmt.Errorf("unsupported socket url: %s (use either grpc:// or ws/wss://)", socketURL)
	}
}

func NewChainFeed(
	iscPackageID iotago.PackageID,
	anchorAddress iotago.ObjectID,
	log log.Logger,
	socketURL string,
	httpURL string,
) (*ChainFeed, error) {
	eventClient, err := selectEventClient(log, socketURL, iscPackageID, anchorAddress)
	if err != nil {
		return nil, err
	}

	httpClient := NewHTTPClient(httpURL, "", iotaclient.WaitForEffectsEnabled)

	return &ChainFeed{
		eventClient:   eventClient,
		httpClient:    httpClient,
		iscPackageID:  iscPackageID,
		anchorAddress: anchorAddress,
		log:           log.NewChildLogger("iscmove-chainfeed"),
	}, nil
}

func (f *ChainFeed) WaitUntilStopped() {
	f.eventClient.WaitUntilStopped()
}

func (f *ChainFeed) GetCurrentAnchor(ctx context.Context) (*iscmove.AnchorWithRef, error) {
	anchor, err := f.httpClient.GetAnchorFromObjectID(ctx, &f.anchorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch anchor: %w", err)
	}
	return anchor, err
}

// FetchCurrentState fetches the current Anchor and all Requests owned by the
// anchor address.
func (f *ChainFeed) FetchCurrentState(ctx context.Context, maxAmountOfRequests int, requestCb func(error, *iscmove.RefWithObject[iscmove.Request])) (*iscmove.AnchorWithRef, error) {
	anchor, err := f.GetCurrentAnchor(ctx)
	if err != nil {
		return nil, err
	}

	// This was refactored from a return based function to a callback based one, as pulling many requests takes
	// a lot of time (~5-10 requests per second) and it would halt ISC on start up, until all requests are pulled.
	// This gives us the option to run this call in a separate goroutine.
	// During my testing I found, that just adding `go` in front of it, isn't enough, and it requires further synchronization from the caller.
	// I kept it as a callback based function for now, as pulling the requests needs improvement and it seems to be the way to go.
	err = f.httpClient.GetRequestsSorted(ctx, f.iscPackageID, &f.anchorAddress, maxAmountOfRequests, requestCb)

	return anchor, err
}

// SubscribeToUpdates starts fetching updated versions of the Anchor and newly received requests in background.
func (f *ChainFeed) SubscribeToUpdates(
	ctx context.Context,
	anchorID iotago.ObjectID,
	anchorCh chan<- *iscmove.AnchorWithRef,
	requestsCh chan<- *iscmove.RefWithObject[iscmove.Request],
) {
	go f.subscribeToAnchorUpdates(ctx, anchorCh)
	go f.subscribeToNewRequests(ctx, anchorID, requestsCh)
}

func (f *ChainFeed) subscribeToNewRequests(
	ctx context.Context,
	anchorID iotago.ObjectID,
	requests chan<- *iscmove.RefWithObject[iscmove.Request],
) {
	for {
		events, err := f.eventClient.SubscribeEvents(ctx)
		if err != nil {
			f.log.LogErrorf("subscribeToNewRequests: err: %s", ctx.Err())
		}

		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
			return
		}

		f.consumeRequestEvents(ctx, events, requests, anchorID)

		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
			return
		}
	}
}

func (f *ChainFeed) consumeRequestEvents(ctx context.Context, events <-chan iscmove.RequestEvent, requests chan<- *iscmove.RefWithObject[iscmove.Request], anchorID iotago.ObjectID) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				continue
			}

			// Drop any request that does not belong to our Anchor (gRPC does not support AND filtering, so we need to do that explicitly here)
			if ev.Anchor != anchorID {
				continue
			}

			f.log.LogDebugf("consumeRequestEvents: received event: requestId: %s, anchorId: %s", ev.RequestID, ev.Anchor)

			reqWithObj, err := f.httpClient.GetRequestFromObjectID(ctx, &ev.RequestID)
			if err != nil {
				f.log.LogErrorf("consumeRequestEvents: cannot fetch Request: %s", err)
				continue
			}

			requests <- reqWithObj

			f.log.LogDebugf("REQUEST[%s] SENT TO CHANNEL %s\n", ev.RequestID.String(), time.Now().String())
		}
	}
}

func (f *ChainFeed) subscribeToAnchorUpdates(
	ctx context.Context,
	anchorCh chan<- *iscmove.AnchorWithRef,
) {
	for {
		changes := make(chan *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects])
		res, err := f.httpClient.QueryTransactionBlocks(ctx, iotaclient.QueryTransactionBlocksRequest{
			Query: &iotajsonrpc.IotaTransactionBlockResponseQuery{
				Filter: &iotajsonrpc.TransactionFilter{
					ChangedObject: &f.anchorAddress,
				},
			}})

		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToAnchorUpdates: ctx.Err(): %s", ctx.Err())
			return
		}

		if err != nil {
			f.log.LogErrorf("subscribeToAnchorUpdates: failed to call SubscribeEvent(): %s", err)
		} else {
			for _, v := range res.Data {
				changes <- v.Effects
			}

			f.consumeAnchorUpdates(ctx, changes, anchorCh)
		}

		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToAnchorUpdates: ctx.Err(): %s", ctx.Err())
			return
		}
	}
}

func (f *ChainFeed) consumeAnchorUpdates(
	ctx context.Context,
	changes <-chan *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects],
	anchorCh chan<- *iscmove.AnchorWithRef,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case change, ok := <-changes:
			if !ok {
				return
			}
			for _, obj := range change.Data.V1.Mutated {
				if *obj.Reference.ObjectID != f.anchorAddress {
					continue
				}

				f.log.LogDebugf("POLLING ANCHOR %s, %s", f.anchorAddress, time.Now().String())

				r, err := f.httpClient.TryGetPastObject(ctx, iotaclient.TryGetPastObjectRequest{
					ObjectID: &f.anchorAddress,
					Version:  obj.Reference.Version,
					Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true, ShowOwner: true, ShowContent: true},
				})
				if err != nil {
					f.log.LogErrorf("consumeAnchorUpdates: cannot fetch Anchor: %s", err)
					continue
				}
				if r.Data.VersionFound == nil {
					f.log.LogErrorf("consumeAnchorUpdates: cannot fetch Anchor: version %d not found", obj.Reference.Version)
					continue
				}

				var anchor *iscmove.Anchor
				err = iotaclient.UnmarshalBCS(r.Data.VersionFound.Bcs.Data.MoveObject.BcsBytes, &anchor)
				if err != nil {
					f.log.LogErrorf("ID: %s\nAssetBagID: %s\n", anchor.ID, anchor.Assets.Value.ID)
					f.log.LogErrorf("consumeAnchorUpdates: failed to unmarshal BCS: %s", err)
					continue
				}

				anchorCh <- &iscmove.AnchorWithRef{
					ObjectRef: r.Data.VersionFound.Ref(),
					Object:    anchor,
					Owner:     r.Data.VersionFound.Owner.AddressOwner,
				}
				f.log.LogDebugf("ANCHOR[%s] SENT TO CHANNEL %s\n", anchor.ID.String(), time.Now().String())
			}
		}
	}
}

func (f *ChainFeed) GetISCPackageID() iotago.PackageID {
	return f.iscPackageID
}

func (f *ChainFeed) GetChainGasCoin(ctx context.Context) (*iotago.ObjectRef, uint64, error) {
	anchor, err := f.httpClient.GetAnchorFromObjectID(ctx, &f.anchorAddress)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch anchor: %w", err)
	}
	metadata, err := transaction.StateMetadataFromBytes(anchor.Object.StateMetadata)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch anchor: %w", err)
	}
	getObjRes, err := f.httpClient.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: metadata.GasCoinObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch gas coin object: %w", err)
	}
	var moveGasCoin MoveCoin
	err = iotaclient.UnmarshalBCS(getObjRes.Data.Bcs.Data.MoveObject.BcsBytes, &moveGasCoin)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode gas coin object: %w", err)
	}
	gasCoinRef := getObjRes.Data.Ref()
	return &gasCoinRef, moveGasCoin.Balance, nil
}
