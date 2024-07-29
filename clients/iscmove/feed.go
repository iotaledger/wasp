package iscmove

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/types"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/serialization"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type ChainFeed struct {
	wsClient      *Client
	iscPackageID  sui.PackageID
	anchorAddress sui.ObjectID
	log           *logger.Logger
}

func NewChainFeed(
	ctx context.Context,
	wsClient *Client,
	iscPackageID sui.PackageID,
	anchorAddress sui.ObjectID,
	log *logger.Logger,
) *ChainFeed {
	return &ChainFeed{
		wsClient:      wsClient,
		iscPackageID:  iscPackageID,
		anchorAddress: anchorAddress,
		log:           log.Named("iscmove-chainfeed"),
	}
}

func (f *ChainFeed) WaitUntilStopped() {
	f.wsClient.WaitUntilStopped()
}

// FetchCurrentState fetches the current Anchor and all Requests owned by the
// anchor address.
func (f *ChainFeed) FetchCurrentState(ctx context.Context) (*types.RefWithObject[types.Anchor], []*types.Request, error) {
	anchor, err := f.wsClient.GetAnchorFromObjectID(ctx, &f.anchorAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch anchor: %w", err)
	}
	reqs := make([]*types.Request, 0)
	var lastSeen *sui.ObjectID
	for {
		res, err := f.wsClient.GetOwnedObjects(ctx, suiclient.GetOwnedObjectsRequest{
			Address: &f.anchorAddress,
			Query: &suijsonrpc.SuiObjectResponseQuery{
				Filter: &suijsonrpc.SuiObjectDataFilter{
					StructType: &sui.StructTag{
						Address: &f.iscPackageID,
						Module:  types.RequestModuleName,
						Name:    types.RequestObjectName,
					},
				},
				Options: &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
			},
			Cursor: lastSeen,
		})
		if ctx.Err() != nil {
			return nil, nil, fmt.Errorf("failed to fetch requests: %w", err)
		}
		if len(res.Data) == 0 {
			break
		}
		lastSeen = res.NextCursor
		for _, reqData := range res.Data {
			var req types.Request
			err := suiclient.UnmarshalBCS(reqData.Data.Bcs.Data.MoveObject.BcsBytes, &req)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode request: %w", err)
			}
			reqs = append(reqs, &req)
		}
	}
	return anchor, reqs, nil
}

// SubscribeToUpdates starts fetching updated versions of the Anchor and newly received requests in background.
func (f *ChainFeed) SubscribeToUpdates(
	ctx context.Context,
	anchorCh chan<- *types.RefWithObject[types.Anchor],
	requestsCh chan<- *types.Request,
) {
	go f.subscribeToAnchorUpdates(ctx, anchorCh)
	go f.subscribeToNewRequests(ctx, requestsCh)
}

func (f *ChainFeed) subscribeToNewRequests(
	ctx context.Context,
	requests chan<- *types.Request,
) {
	for {
		events := make(chan *suijsonrpc.SuiEvent)
		err := f.wsClient.SubscribeEvent(
			ctx,
			&suijsonrpc.EventFilter{
				MoveEventType: &sui.StructTag{
					Address: &f.iscPackageID,
					Module:  types.RequestModuleName,
					Name:    types.RequestEventObjectName,
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
	events <-chan *suijsonrpc.SuiEvent,
	requests chan<- *types.Request,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			var reqEvent types.RequestEvent
			err := suiclient.UnmarshalBCS(ev.Bcs, &reqEvent)
			if err != nil {
				f.log.Errorf("consumeRequestEvents: cannot decode RequestEvent BCS: %s", err)
				continue
			}
			req, err := f.wsClient.GetRequestFromObjectID(ctx, &reqEvent.RequestID)
			if err != nil {
				f.log.Errorf("consumeRequestEvents: cannot fetch Request: %s", err)
				continue
			}
			requests <- req
		}
	}
}

func (f *ChainFeed) subscribeToAnchorUpdates(
	ctx context.Context,
	anchorCh chan<- *types.RefWithObject[types.Anchor],
) {
	for {
		changes := make(chan *serialization.TagJson[suijsonrpc.SuiTransactionBlockEffects])
		err := f.wsClient.SubscribeTransaction(
			ctx,
			&suijsonrpc.TransactionFilter{
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
	changes <-chan *serialization.TagJson[suijsonrpc.SuiTransactionBlockEffects],
	anchorCh chan<- *types.RefWithObject[types.Anchor],
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
				if *obj.Reference.ObjectID == f.anchorAddress {
					r, err := f.wsClient.TryGetPastObject(ctx, suiclient.TryGetPastObjectRequest{
						ObjectID: &f.anchorAddress,
						Version:  obj.Reference.Version,
						Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
					})
					if err != nil {
						f.log.Errorf("consumeAnchorUpdates: cannot fetch Anchor: %s", err)
						continue
					}
					if r.Data.VersionFound == nil {
						f.log.Errorf("consumeAnchorUpdates: cannot fetch Anchor: version %d not found", obj.Reference.Version)
						continue
					}
					var anchor *types.Anchor
					err = suiclient.UnmarshalBCS(r.Data.VersionFound.Bcs.Data.MoveObject.BcsBytes, &anchor)
					if err != nil {
						f.log.Errorf("consumeAnchorUpdates: failed to unmarshal BCS: %s", err)
						continue
					}
					anchorCh <- &types.RefWithObject[types.Anchor]{
						ObjectRef: r.Data.VersionFound.Ref(),
						Object:    anchor,
					}
				}
			}
		}
	}
}
