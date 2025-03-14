package iscmoveclient

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/transaction"
)

type ChainFeed struct {
	wsClient      *Client
	httpClient    *Client
	iscPackageID  iotago.PackageID
	anchorAddress iotago.ObjectID
	log           log.Logger
}

func NewChainFeed(
	ctx context.Context,
	iscPackageID iotago.PackageID,
	anchorAddress iotago.ObjectID,
	log log.Logger,
	wsURL string,
	httpURL string,
) (*ChainFeed, error) {
	wsClient, err := NewWebsocketClient(ctx, wsURL, "", iotaclient.WaitForEffectsEnabled, log)
	if err != nil {
		return nil, err
	}

	httpClient := NewHTTPClient(httpURL, "", iotaclient.WaitForEffectsEnabled)

	return &ChainFeed{
		wsClient:      wsClient,
		httpClient:    httpClient,
		iscPackageID:  iscPackageID,
		anchorAddress: anchorAddress,
		log:           log.NewChildLogger("iscmove-chainfeed"),
	}, nil
}

func (f *ChainFeed) WaitUntilStopped() {
	f.wsClient.WaitUntilStopped()
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
		events := make(chan *iotajsonrpc.IotaEvent)
		err := f.wsClient.SubscribeEvent(
			ctx,
			&iotajsonrpc.EventFilter{
				And: &iotajsonrpc.AndOrEventFilter{
					Filter1: &iotajsonrpc.EventFilter{MoveEventType: &iotago.StructTag{
						Address: &f.iscPackageID,
						Module:  iscmove.RequestModuleName,
						Name:    iscmove.RequestEventObjectName,
					}},
					Filter2: &iotajsonrpc.EventFilter{MoveEventField: &iotajsonrpc.EventFilterMoveEventField{
						Path:  iscmove.RequestEventAnchorFieldName,
						Value: anchorID.String(),
					}},
				},
			},
			events,
		)
		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
			return
		}
		if err != nil {
			f.log.LogErrorf("subscribeToNewRequests: failed to call SubscribeEvent(): %s", err)
		} else {
			f.consumeRequestEvents(ctx, events, requests)
		}
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
			return
		}
	}
}

func (f *ChainFeed) consumeRequestEvents(
	ctx context.Context,
	events <-chan *iotajsonrpc.IotaEvent,
	requests chan<- *iscmove.RefWithObject[iscmove.Request],
) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			var reqEvent iscmove.RequestEvent
			err := iotaclient.UnmarshalBCS(ev.Bcs, &reqEvent)
			if err != nil {
				f.log.LogErrorf("consumeRequestEvents: cannot decode RequestEvent BCS: %s", err)
				continue
			}

			reqWithObj, err := f.httpClient.GetRequestFromObjectID(ctx, &reqEvent.RequestID)
			if err != nil {
				f.log.LogErrorf("consumeRequestEvents: cannot fetch Request: %s", err)
				continue
			}

			requests <- reqWithObj

			f.log.LogDebugf("REQUEST[%s] SENT TO CHANNEL %s\n", reqEvent.RequestID.String(), time.Now().String())
		}
	}
}

func (f *ChainFeed) subscribeToAnchorUpdates(
	ctx context.Context,
	anchorCh chan<- *iscmove.AnchorWithRef,
) {
	for {
		changes := make(chan *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects])
		err := f.wsClient.SubscribeTransaction(
			ctx,
			&iotajsonrpc.TransactionFilter{
				ChangedObject: &f.anchorAddress,
			},
			changes,
		)
		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToAnchorUpdates: ctx.Err(): %s", ctx.Err())
			return
		}
		if err != nil {
			f.log.LogErrorf("subscribeToAnchorUpdates: failed to call SubscribeEvent(): %s", err)
		} else {
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
