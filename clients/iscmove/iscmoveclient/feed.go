package iscmoveclient

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/transaction"
)

type ChainFeed struct {
	wsClient               *Client // FIXME this should be removed after we migrate to GqraphQL subscriptions
	httpClient             *Client
	iscPackageID           iotago.PackageID
	anchorAddress          iotago.ObjectID
	anchorFetchMaxAttempts int
	anchorFetchRetryDelay  time.Duration
	log                    log.Logger
}

func NewChainFeed(
	ctx context.Context,
	iscPackageID iotago.PackageID,
	anchorAddress iotago.ObjectID,
	log log.Logger,
	wsURL string,
	httpURL string,
	anchorFetchMaxAttempts int,
	anchorFetchRetryDelay time.Duration,
) (*ChainFeed, error) {
	graphqlLog := log.NewChildLogger("graphql")
	wsGQL := iotagraphql.NewGraphQLClientWithWaitParams(wsURL, "", iotagraphql.WaitForEffectsEnabled).WithLogger(graphqlLog)
	wsClient := NewClient(wsGQL)

	httpGQL := iotagraphql.NewGraphQLClientWithWaitParams(httpURL, "", iotagraphql.WaitForEffectsEnabled).WithLogger(graphqlLog)
	httpClient := NewClient(httpGQL)

	return &ChainFeed{
		wsClient:               wsClient,
		httpClient:             httpClient,
		iscPackageID:           iscPackageID,
		anchorAddress:          anchorAddress,
		anchorFetchMaxAttempts: anchorFetchMaxAttempts,
		anchorFetchRetryDelay:  anchorFetchRetryDelay,
		log:                    log.NewChildLogger("iscmove-chainfeed"),
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
// signerAddress is the committee address that signs transactions updating the anchor.
func (f *ChainFeed) SubscribeToUpdates(
	ctx context.Context,
	anchorID iotago.ObjectID,
	signerAddress iotago.Address,
	anchorCh chan<- *iscmove.AnchorWithRef,
	requestsCh chan<- *iscmove.RefWithObject[iscmove.Request],
) {
	go f.subscribeToAnchorUpdates(ctx, signerAddress, anchorCh)
	go f.subscribeToNewRequests(ctx, anchorID, requestsCh)
}

func (f *ChainFeed) subscribeToNewRequests(
	ctx context.Context,
	anchorID iotago.ObjectID,
	requests chan<- *iscmove.RefWithObject[iscmove.Request],
) {
	for {
		events := make(chan *iotagraphql.IotaEvent)
		err := f.wsClient.SubscribeEvent(
			ctx,
			&iotagraphql.IotaEventFilter{
				MoveModule: &iotagraphql.IotaEventFilterMoveModule{
					Package: &f.iscPackageID,
					Module:  string(iscmove.RequestModuleName),
				},
				MoveEventType: &iotago.StructTag{
					Address: &f.iscPackageID,
					Module:  iscmove.RequestModuleName,
					Name:    iscmove.RequestEventObjectName,
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
			f.consumeRequestEvents(ctx, events, requests, anchorID)
		}
		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToNewRequests: ctx.Err(): %s", ctx.Err())
			return
		}
	}
}

func (f *ChainFeed) consumeRequestEvents(
	ctx context.Context,
	events <-chan *iotagraphql.IotaEvent,
	requests chan<- *iscmove.RefWithObject[iscmove.Request],
	anchorID iotago.ObjectID,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			f.log.LogDebugf("consumeRequestEvents: received request event: %+v", ev)
			var reqEvent iscmove.RequestEvent
			err := iotagraphql.UnmarshalBCS(ev.Bcs, &reqEvent)
			if err != nil {
				f.log.LogErrorf("consumeRequestEvents: cannot decode RequestEvent BCS: %s", err)
				continue
			}

			// skip if event is not from current anchor
			f.log.LogDebugf("consumeRequestEvents: anchorID: %s, reqEvent.Anchor: %s", anchorID.String(), reqEvent.Anchor.String())
			if reqEvent.Anchor != anchorID {
				f.log.LogDebugf("consumeRequestEvents: skipping request event for different anchor: %s", reqEvent.Anchor.String())
				continue
			}

			f.log.LogDebugf("consumeRequestEvents: fetching request: %s", reqEvent.RequestID.String())

			reqWithObj, err := f.httpClient.GetRequestFromObjectID(ctx, &reqEvent.RequestID)
			if err != nil {
				f.log.LogErrorf("consumeRequestEvents: cannot fetch Request: %s", err)
				continue
			}

			f.log.LogDebugf("consumeRequestEvents: sending request to channel: %+v", reqWithObj)
			requests <- reqWithObj

			f.log.LogDebugf("REQUEST[%s] SENT TO CHANNEL %s\n", reqEvent.RequestID.String(), time.Now().String())
		}
	}
}

func (f *ChainFeed) subscribeToAnchorUpdates(
	ctx context.Context,
	signerAddress iotago.Address,
	anchorCh chan<- *iscmove.AnchorWithRef,
) {
	for {
		changes := make(chan *iotagraphql.IotaTransactionBlockEffects)
		err := f.wsClient.SubscribeTransaction(
			ctx,
			&iotagraphql.TransactionFilter{
				FromAddress:   &signerAddress,
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
		if ctx.Err() != nil {
			f.log.LogErrorf("subscribeToAnchorUpdates: ctx.Err(): %s", ctx.Err())
			return
		}
	}
}

func (f *ChainFeed) consumeAnchorUpdates(
	ctx context.Context,
	changes <-chan *iotagraphql.IotaTransactionBlockEffects,
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
			f.log.LogDebugf("consumeAnchorUpdates: received anchor update: %+v", change)
			for _, obj := range change.V1.Mutated {
				if *obj.Reference.ObjectID != f.anchorAddress {
					continue
				}

				f.log.LogDebugf("POLLING ANCHOR %s, %s", f.anchorAddress, time.Now().String())

				anchorWithRef, err := f.fetchAnchorWithRetry(ctx, obj.Reference.Version)
				if err != nil {
					f.log.LogErrorf("consumeAnchorUpdates: giving up on anchor version %d: %s", obj.Reference.Version, err)
					continue
				}

				anchorCh <- anchorWithRef
				f.log.LogDebugf("ANCHOR[%s] SENT TO CHANNEL %s\n", anchorWithRef.Object.ID.String(), time.Now().String())
			}
		}
	}
}


func (f *ChainFeed) fetchAnchorWithRetry(ctx context.Context, version uint64) (*iscmove.AnchorWithRef, error) {
	for attempt := range f.anchorFetchMaxAttempts {
		r, err := f.httpClient.TryGetPastObject(ctx, f.anchorAddress, version)
		if err != nil {
			f.log.LogDebugf("fetchAnchorWithRetry: attempt %d/%d failed: %s", attempt+1, f.anchorFetchMaxAttempts, err)
		} else if r.Object.IsNotFound() {
			f.log.LogDebugf("fetchAnchorWithRetry: attempt %d/%d version %d not found", attempt+1, f.anchorFetchMaxAttempts, version)
		} else {
			var anchor *iscmove.Anchor
			err = iotagraphql.UnmarshalBCS(r.Object.BcsBytes(), &anchor)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal anchor BCS: %w", err)
			}
			objRef, err := r.Object.ObjectRef()
			if err != nil {
				return nil, fmt.Errorf("failed to get object ref: %w", err)
			}

			f.log.LogDebugf("fetchAnchorWithRetry: found anchor after %d/%d attempts.", attempt+1, f.anchorFetchMaxAttempts)

			return &iscmove.AnchorWithRef{
				ObjectRef: *objRef,
				Object:    anchor,
				Owner:     r.Object.OwnerAddress(),
			}, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(f.anchorFetchRetryDelay):
		}
	}
	return nil, fmt.Errorf("anchor version %d not available after %d attempts", version, f.anchorFetchMaxAttempts)
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
	getObjRes, err := f.httpClient.GetObject(ctx, *metadata.GasCoinObjectID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch gas coin object: %w", err)
	}
	var moveGasCoin MoveCoin
	err = iotagraphql.UnmarshalBCS(getObjRes.Object.BcsBytes(), &moveGasCoin)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode gas coin object: %w", err)
	}
	gasCoinRef, err := getObjRes.Object.ObjectRef()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get gas coin ref: %w", err)
	}
	return gasCoinRef, moveGasCoin.Balance, nil
}
