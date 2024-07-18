package iscmove

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/serialization"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type Feed struct {
	Sui           *suiclient.WebsocketClient
	ISC           *Client // TODO: remove this, should be able to use the same websocket connection for all calls
	ISCPackageID  sui.PackageID
	AnchorAddress sui.ObjectID
	log           *logger.Logger
}

func NewRequestsFeed(
	config Config,
	wsURL string,
	iscPackageID sui.PackageID,
	anchorAddress sui.ObjectID,
	log *logger.Logger,
) *Feed {
	return &Feed{
		Sui:           suiclient.NewWebsocket(config.APIURL, wsURL),
		ISC:           NewClient(config),
		ISCPackageID:  iscPackageID,
		AnchorAddress: anchorAddress,
		log:           log,
	}
}

// FetchCurrentState fetches the current Anchor and all Requests owned by the
// anchor address.
func (c *Feed) FetchCurrentState(ctx context.Context) (*RefWithObject[Anchor], []*Request, error) {
	anchor, err := c.ISC.GetAnchorFromObjectID(ctx, &c.AnchorAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch anchor: %w", err)
	}
	reqs := make([]*Request, 0)
	var lastSeen *sui.ObjectID
	for {
		res, err := c.Sui.GetOwnedObjects(ctx, suiclient.GetOwnedObjectsRequest{
			Address: &c.AnchorAddress,
			Query: &suijsonrpc.SuiObjectResponseQuery{
				Filter: &suijsonrpc.SuiObjectDataFilter{
					StructType: &sui.StructTag{
						Address: &c.ISCPackageID,
						Module:  RequestModuleName,
						Name:    RequestObjectName,
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
			var req Request
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
func (c *Feed) SubscribeToUpdates(
	ctx context.Context,
	anchorCh chan<- *RefWithObject[Anchor],
	requestsCh chan<- *Request,
) {
	// TODO fix
	// go c.subscribeToAnchorUpdates(ctx, anchorCh)
	go c.subscribeToNewRequests(ctx, requestsCh)
}

func (c *Feed) subscribeToNewRequests(
	ctx context.Context,
	requests chan<- *Request,
) {
	for {
		events := make(chan *suijsonrpc.SuiEvent)
		err := c.Sui.SubscribeEvent(
			ctx,
			&suijsonrpc.EventFilter{
				MoveEventType: &sui.StructTag{
					Address: &c.ISCPackageID,
					Module:  RequestModuleName,
					Name:    RequestEventObjectName,
				},
			},
			events,
		)
		if ctx.Err() != nil {
			c.log.Errorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
			return
		}
		if err != nil {
			c.log.Errorf("subscribeToNewRequests: failed to call SubscribeEvent(): %s", err)
		} else {
			c.consumeRequestEvents(ctx, events, requests)
		}
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			c.log.Errorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
			return
		}
	}
}

func (c *Feed) consumeRequestEvents(
	ctx context.Context,
	events <-chan *suijsonrpc.SuiEvent,
	requests chan<- *Request,
) {
	for {
		select {
		case <-ctx.Done():
			break
		case ev, ok := <-events:
			if !ok {
				break
			}
			var reqEvent RequestEvent
			err := suiclient.UnmarshalBCS(ev.Bcs, &reqEvent)
			if err != nil {
				c.log.Errorf("consumeRequestEvents: cannot decode RequestEvent BCS: %s", err)
				continue
			}
			req, err := c.ISC.GetRequestFromObjectID(ctx, &reqEvent.RequestID)
			if err != nil {
				c.log.Errorf("consumeRequestEvents: cannot fetch Request: %s", err)
				continue
			}
			requests <- req
		}
	}
}

func (c *Feed) subscribeToAnchorUpdates(
	ctx context.Context,
	anchorCh chan<- *RefWithObject[Anchor],
) {
	for {
		changes := make(chan *serialization.TagJson[suijsonrpc.SuiTransactionBlockEffects])
		err := c.Sui.SubscribeTransaction(
			ctx,
			&suijsonrpc.TransactionFilter{
				ChangedObject: &c.AnchorAddress,
			},
			changes,
		)
		if ctx.Err() != nil {
			c.log.Errorf("subscribeToAnchorUpdates: ctx.Err(): %s", ctx.Err())
			return
		}
		if err != nil {
			c.log.Errorf("subscribeToAnchorUpdates: failed to call SubscribeEvent(): %s", err)
		} else {
			c.consumeAnchorUpdates(ctx, changes, anchorCh)
		}
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			c.log.Errorf("subscribeToAnchorUpdates: ctx.Err(): %s", ctx.Err())
			return
		}
	}
}

func (c *Feed) consumeAnchorUpdates(
	ctx context.Context,
	changes <-chan *serialization.TagJson[suijsonrpc.SuiTransactionBlockEffects],
	anchorCh chan<- *RefWithObject[Anchor],
) {
	for {
		select {
		case <-ctx.Done():
			break
		case change, ok := <-changes:
			if !ok {
				break
			}
			for _, obj := range change.Data.V1.Mutated {
				if *obj.Reference.ObjectID == c.AnchorAddress {
					r, err := c.Sui.TryGetPastObject(ctx, suiclient.TryGetPastObjectRequest{
						ObjectID: &c.AnchorAddress,
						Version:  obj.Reference.Version,
						Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
					})
					if err != nil {
						c.log.Errorf("consumeAnchorUpdates: cannot fetch Anchor: %s", err)
						continue
					}
					if r.Data.VersionFound == nil {
						c.log.Errorf("consumeAnchorUpdates: cannot fetch Anchor: version %d not found", obj.Reference.Version)
						continue
					}
					var anchor *Anchor
					err = suiclient.UnmarshalBCS(r.Data.VersionFound.Bcs.Data.MoveObject.BcsBytes, &anchor)
					if err != nil {
						c.log.Errorf("consumeAnchorUpdates: failed to unmarshal BCS: %s", err)
						continue
					}
					anchorCh <- &RefWithObject[Anchor]{
						ObjectRef: r.Data.VersionFound.Ref(),
						Object:    anchor,
					}
				}
			}
		}
	}
}
