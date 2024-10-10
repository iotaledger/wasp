package iscmoveclient

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/clients/iscmove"
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/sui/serialization"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	suijsonrpc2 "github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
)

type ChainFeed struct {
	wsClient      *Client
	iscPackageID  sui2.PackageID
	anchorAddress sui2.ObjectID
	log           *logger.Logger
}

func NewChainFeed(
	ctx context.Context,
	wsClient *Client,
	iscPackageID sui2.PackageID,
	anchorAddress sui2.ObjectID,
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
func (f *ChainFeed) FetchCurrentState(ctx context.Context) (*iscmove.AnchorWithRef, []*iscmove.RefWithObject[iscmove.Request], error) {
	anchor, err := f.wsClient.GetAnchorFromObjectID(ctx, &f.anchorAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch anchor: %w", err)
	}
	reqs := make([]*iscmove.RefWithObject[iscmove.Request], 0)
	var lastSeen *sui2.ObjectID
	for {
		res, err := f.wsClient.GetOwnedObjects(ctx, suiclient2.GetOwnedObjectsRequest{
			Address: &f.anchorAddress,
			Query: &suijsonrpc2.SuiObjectResponseQuery{
				Filter: &suijsonrpc2.SuiObjectDataFilter{
					StructType: &sui2.StructTag{
						Address: &f.iscPackageID,
						Module:  iscmove.RequestModuleName,
						Name:    iscmove.RequestObjectName,
					},
				},
				Options: &suijsonrpc2.SuiObjectDataOptions{ShowBcs: true},
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
			var req moveRequest
			err := suiclient2.UnmarshalBCS(reqData.Data.Bcs.Data.MoveObject.BcsBytes, &req)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode request: %w", err)
			}
			bals, err := f.wsClient.GetAssetsBagWithBalances(ctx, &req.assetsBag.ID)
			if ctx.Err() != nil {
				return nil, nil, fmt.Errorf("failed to fetch AssetsBag of Request: %w", ctx.Err())
			}
			if err != nil {
				return nil, nil, fmt.Errorf("failed to fetch AssetsBag of Request: %w", err)
			}
			req.assetsBag.Value = bals
			reqs = append(reqs, &iscmove.RefWithObject[iscmove.Request]{
				ObjectRef: reqData.Data.Ref(),
				Object:    req.ToRequest(),
			})
		}
	}
	return anchor, reqs, nil
}

// SubscribeToUpdates starts fetching updated versions of the Anchor and newly received requests in background.
func (f *ChainFeed) SubscribeToUpdates(
	ctx context.Context,
	anchorID sui2.ObjectID,
	anchorCh chan<- *iscmove.AnchorWithRef,
	requestsCh chan<- *iscmove.RefWithObject[iscmove.Request],
) {
	go f.subscribeToAnchorUpdates(ctx, anchorCh)
	go f.subscribeToNewRequests(ctx, anchorID, requestsCh)
}

func (f *ChainFeed) subscribeToNewRequests(
	ctx context.Context,
	anchorID sui2.ObjectID,
	requests chan<- *iscmove.RefWithObject[iscmove.Request],
) {
	for {
		events := make(chan *suijsonrpc2.SuiEvent)
		err := f.wsClient.SubscribeEvent(
			ctx,
			&suijsonrpc2.EventFilter{
				And: &suijsonrpc2.AndOrEventFilter{
					Filter1: &suijsonrpc2.EventFilter{MoveEventType: &sui2.StructTag{
						Address: &f.iscPackageID,
						Module:  iscmove.RequestModuleName,
						Name:    iscmove.RequestEventObjectName,
					}},
					Filter2: &suijsonrpc2.EventFilter{MoveEventField: &suijsonrpc2.EventFilterMoveEventField{
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
	events <-chan *suijsonrpc2.SuiEvent,
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
			err := suiclient2.UnmarshalBCS(ev.Bcs, &reqEvent)
			if err != nil {
				f.log.Errorf("consumeRequestEvents: cannot decode RequestEvent BCS: %s", err)
				continue
			}
			reqWithObj, err := f.wsClient.GetRequestFromObjectID(ctx, &reqEvent.RequestID)
			if err != nil {
				f.log.Errorf("consumeRequestEvents: cannot fetch Request: %s", err)
				continue
			}
			requests <- reqWithObj
		}
	}
}

func (f *ChainFeed) subscribeToAnchorUpdates(
	ctx context.Context,
	anchorCh chan<- *iscmove.AnchorWithRef,
) {
	for {
		changes := make(chan *serialization.TagJson[suijsonrpc2.SuiTransactionBlockEffects])
		err := f.wsClient.SubscribeTransaction(
			ctx,
			&suijsonrpc2.TransactionFilter{
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
	changes <-chan *serialization.TagJson[suijsonrpc2.SuiTransactionBlockEffects],
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
				if *obj.Reference.ObjectID == f.anchorAddress {
					r, err := f.wsClient.TryGetPastObject(ctx, suiclient2.TryGetPastObjectRequest{
						ObjectID: &f.anchorAddress,
						Version:  obj.Reference.Version,
						Options:  &suijsonrpc2.SuiObjectDataOptions{ShowBcs: true},
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
					err = suiclient2.UnmarshalBCS(r.Data.VersionFound.Bcs.Data.MoveObject.BcsBytes, &anchor)
					if err != nil {
						f.log.Errorf("consumeAnchorUpdates: failed to unmarshal BCS: %s", err)
						continue
					}
					anchorCh <- &iscmove.AnchorWithRef{
						ObjectRef: r.Data.VersionFound.Ref(),
						Object:    anchor,
					}
				}
			}
		}
	}
}

func (f *ChainFeed) GetISCPackageID() sui2.PackageID {
	return f.iscPackageID
}
