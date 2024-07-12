package iscmove

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type RequestsFeed struct {
	Sui           *suiclient.WebsocketClient
	ISCPackageID  sui.PackageID
	AnchorAddress sui.ObjectID
	log           *logger.Logger
}

func NewRequestsFeed(
	apiURL string,
	wsURL string,
	iscPackageID sui.PackageID,
	anchorAddress sui.ObjectID,
	log *logger.Logger,
) *RequestsFeed {
	return &RequestsFeed{
		Sui:           suiclient.NewWebsocket(apiURL, wsURL),
		ISCPackageID:  iscPackageID,
		AnchorAddress: anchorAddress,
		log:           log,
	}
}

// SubscribeOwnedRequests fetches requests owned by the anchor address in background.
// The given channel will yield a new item for each Request currently owned by
// the anchor address. When all requests have been fetched, the channel is
// closed.
func (c *RequestsFeed) SubscribeOwnedRequests(
	ctx context.Context,
	requests chan<- Request,
) {
	go c.fetchAllRequestsOwned(ctx, requests)
}

// SubscribeNewRequests starts fetching newly received requests in background.
// The given channel will yield a new Request every time the RequestEvent is
// emitted by the Anchor object.
func (c *RequestsFeed) SubscribeNewRequests(
	ctx context.Context,
	requests chan<- Request,
) {
	go c.subscribeToNewRequests(ctx, requests)
}

func (c *RequestsFeed) fetchAllRequestsOwned(
	ctx context.Context,
	requests chan<- Request,
) {
	var lastSeen *sui.ObjectID
	for {
		res, err := c.Sui.GetOwnedObjects(ctx, suiclient.GetOwnedObjectsRequest{
			Address: &c.AnchorAddress,
			Query: &suijsonrpc.SuiObjectResponseQuery{
				Filter: &suijsonrpc.SuiObjectDataFilter{
					StructType: (&sui.StructTag{
						Address: &c.ISCPackageID,
						Module:  RequestModuleName,
						Name:    RequestObjectName,
					}).String(),
				},
				Options: &suijsonrpc.SuiObjectDataOptions{
					ShowBcs: true,
				},
			},
			Cursor: lastSeen,
		})
		if ctx.Err() != nil {
			c.log.Errorf("fetchAllRequestsOwned: ctx.Err(): %s", err)
			return
		}
		if err != nil {
			c.log.Errorf("fetchAllRequestsOwned: failed to call GetOwnedObjects(): %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if len(res.Data) == 0 {
			break
		}
		lastSeen = res.NextCursor
		for _, reqData := range res.Data {
			var req Request
			err := suiclient.UnmarshalBCS(reqData.Data.Bcs.Data.MoveObject.BcsBytes, &req)
			if err != nil {
				c.log.Errorf("fetchAllRequestsOwned: UnmarshalBCS(): %s", err)
				continue
			}
			requests <- req
		}
	}

	// signal that fetching requests is done
	close(requests)
}

func (c *RequestsFeed) subscribeToNewRequests(
	ctx context.Context,
	requests chan<- Request,
) {
	for {
		events := make(chan suijsonrpc.SuiEvent)
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

func (c *RequestsFeed) consumeRequestEvents(
	ctx context.Context,
	events <-chan suijsonrpc.SuiEvent,
	requests chan<- Request,
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
			getObjectRes, err := c.Sui.GetObject(ctx, suiclient.GetObjectRequest{
				ObjectID: &reqEvent.RequestID,
				Options: &suijsonrpc.SuiObjectDataOptions{
					ShowBcs: true,
				},
			})
			if err != nil {
				c.log.Errorf("consumeRequestEvents: cannot fetch Request: %s", err)
				continue
			}
			var req Request
			err = suiclient.UnmarshalBCS(getObjectRes.Data.Bcs.Data.MoveObject.BcsBytes, &req)
			if err != nil {
				c.log.Errorf("consumeRequestEvents: cannot decode Request BCS: %s", err)
				continue
			}
			requests <- req
		}
	}
}
