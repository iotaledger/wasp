package iscmoveclient

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/transaction"
)

type ChainFeed struct {
	wsClient      *Client
	iscPackageID  iotago.PackageID
	anchorAddress iotago.ObjectID
	log           *logger.Logger
}

func NewChainFeed(
	ctx context.Context,
	wsClient *Client,
	iscPackageID iotago.PackageID,
	anchorAddress iotago.ObjectID,
	log *logger.Logger,
) *ChainFeed {
	return &ChainFeed{
		wsClient:      wsClient,
		iscPackageID:  iscPackageID,
		anchorAddress: anchorAddress,
		log:           log.Named("iscmove-chainfeed"),
	}
}

func (f *ChainFeed) ConnectionRecreated() <-chan struct{} {
	return f.wsClient.ConnectionRecreated()
}

func (f *ChainFeed) WaitUntilStopped() {
	f.wsClient.WaitUntilStopped()
}

// FetchCurrentState fetches the current Anchor and all Requests owned by the
// anchor address.
func (f *ChainFeed) FetchCurrentState(ctx context.Context) (*iscmove.AnchorWithRef, []*iscmove.RefWithObject[iscmove.Request], error) {
	anchor, err := f.wsClient.GetAnchorFromObjectID(ctx, &f.anchorAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch anchor: %w", err)
	}

	reqs, err := f.wsClient.GetRequests(ctx, f.iscPackageID, &f.anchorAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch requests: %w", err)
	}

	return anchor, reqs, nil
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
	f.log.Info("subscribeToNewRequests: loop started")
	defer f.log.Info("subscribeToNewRequests: loop exited")
	defer close(requests)

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
			f.log.Errorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
			return
		}
		if err != nil {
			f.log.Errorf("subscribeToNewRequests: failed to call SubscribeEvent(): %s", err)
		} else {
			f.consumeRequestEvents(ctx, events, requests)
		}
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			f.log.Errorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
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
		if ctx.Err() != nil {
			f.log.Info("consumeRequestEvents: context done")
			return
		}
		select {
		case <-ctx.Done():
			f.log.Info("consumeRequestEvents: context done")
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			var reqEvent iscmove.RequestEvent
			err := iotaclient.UnmarshalBCS(ev.Bcs, &reqEvent)
			if err != nil {
				f.log.Errorf("consumeRequestEvents: cannot decode RequestEvent BCS: %s", err)
				continue
			}

			reqWithObj, err := f.wsClient.GetRequestFromObjectID(ctx, &reqEvent.RequestID)
			if err != nil {
				f.log.Errorf("consumeRequestEvents: cannot fetch Request: %s", err)
				continue
			}

			select {
			case <-ctx.Done():
				f.log.Info("consumeRequestEvents: context done")
				return
			case requests <- reqWithObj:
			}
		}
	}
}

func (f *ChainFeed) subscribeToAnchorUpdates(
	ctx context.Context,
	anchorCh chan<- *iscmove.AnchorWithRef,
) {
	f.log.Info("subscribeToAnchorUpdates: loop started")
	defer f.log.Info("subscribeToAnchorUpdates: loop exited")
	defer close(anchorCh)

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
			f.log.Errorf("subscribeToAnchorUpdates: ctx.Err(): %s", ctx.Err())
			return
		}
		if err != nil {
			f.log.Errorf("subscribeToAnchorUpdates: failed to call SubscribeEvent(): %s", err)
		} else {
			f.consumeAnchorUpdates(ctx, changes, anchorCh)
		}
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			f.log.Errorf("subscribeToAnchorUpdates: ctx.Err(): %s", ctx.Err())
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
		if ctx.Err() != nil {
			f.log.Info("consumeRequestEvents: context done")
			return
		}

		select {
		case <-ctx.Done():
			f.log.Info("consumeAnchorUpdates: context done")
			return
		case change, ok := <-changes:
			if !ok {
				return
			}
			for _, obj := range change.Data.V1.Mutated {
				if *obj.Reference.ObjectID == f.anchorAddress {
					r, err := f.wsClient.TryGetPastObject(ctx, iotaclient.TryGetPastObjectRequest{
						ObjectID: &f.anchorAddress,
						Version:  obj.Reference.Version,
						Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true, ShowOwner: true, ShowContent: true},
					})
					if err != nil {
						f.log.Errorf("consumeAnchorUpdates: cannot fetch Anchor: %s", err)
						continue
					}
					if r.Data.VersionFound == nil {
						f.log.Errorf("consumeAnchorUpdates: cannot fetch Anchor: version %d not found", obj.Reference.Version)
						continue
					}
					var anchor *iscmove.Anchor
					err = iotaclient.UnmarshalBCS(r.Data.VersionFound.Bcs.Data.MoveObject.BcsBytes, &anchor)
					if err != nil {
						f.log.Errorf("ID: %s\nAssetBagID: %s\n", anchor.ID, anchor.Assets.Value.ID)
						f.log.Errorf("consumeAnchorUpdates: failed to unmarshal BCS: %s", err)
						continue
					}

					select {
					case <-ctx.Done():
						f.log.Info("consumeAnchorUpdates: context done")
						return
					case anchorCh <- &iscmove.AnchorWithRef{
						ObjectRef: r.Data.VersionFound.Ref(),
						Object:    anchor,
						Owner:     r.Data.VersionFound.Owner.AddressOwner,
					}:
					}
				}
			}
		}
	}
}

func (f *ChainFeed) GetISCPackageID() iotago.PackageID {
	return f.iscPackageID
}

func (f *ChainFeed) GetChainGasCoin(ctx context.Context) (*iotago.ObjectRef, uint64, error) {
	anchor, err := f.wsClient.GetAnchorFromObjectID(ctx, &f.anchorAddress)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch anchor: %w", err)
	}
	metadata, err := transaction.StateMetadataFromBytes(anchor.Object.StateMetadata)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch anchor: %w", err)
	}
	getObjRes, err := f.wsClient.GetObject(ctx, iotaclient.GetObjectRequest{
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
