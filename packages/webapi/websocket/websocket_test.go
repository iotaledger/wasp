package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	websocketserver "nhooyr.io/websocket"

	"github.com/iotaledger/hive.go/core/configuration"
	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/solo"
)

func InitWebsocket(ctx context.Context, t *testing.T, eventsToSubscribe []publisher.ISCEventType) (*Service, *websockethub.Hub, *solo.Chain, *publisher.Publisher) {
	logger.InitGlobalLogger(configuration.New())
	log := logger.NewLogger("Test")

	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Log: log})

	websocketHub := websockethub.NewHub(log.Named("Hub"), &websocketserver.AcceptOptions{InsecureSkipVerify: true}, 500, 500, 500)
	ws := NewWebsocketService(log.Named("Service"), websocketHub, []publisher.ISCEventType{
		publisher.ISCEventKindNewBlock,
		publisher.ISCEventKindReceipt,
		publisher.ISCEventKindBlockEvents,
		publisher.ISCEventIssuerVM,
	}, env.Publisher())

	ws.subscriptionManager.Connect(websockethub.ClientID(0))
	ws.subscriptionManager.Subscribe(websockethub.ClientID(0), "chains")

	for _, eventType := range eventsToSubscribe {
		ws.subscriptionManager.Subscribe(websockethub.ClientID(0), string(eventType))
	}

	ws.EventHandler().AttachToEvents()

	chain := env.NewChain()

	go func() {
		env.Publisher().Run(ctx)
	}()

	go func() {
		websocketHub.Run(ctx)
	}()

	return ws, websocketHub, chain, env.Publisher()
}

func TestWebsocketEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	ws, _, chain, pub := InitWebsocket(ctx, t, []publisher.ISCEventType{publisher.ISCEventKindNewBlock})

	// publisherEvent is the event that gets called from the websocket event handlers.
	//
	//		publisher -> TEvent(ISCEvent[T]) -> Websocket Service EventHandlers
	//			-> Mapping -> publisherEvent(ISCEvent) -> Websocket Send
	//
	// It's the last step before the events get send via the websocket to the client.
	// It's also the last step to validate the events without actually connecting with a websocket client.
	ws.publisherEvent.Attach(event.NewClosure(func(iscEvent *ISCEvent) {
		require.Exactly(t, iscEvent.ChainID, chain.ChainID.String())

		if iscEvent.Kind == publisher.ISCEventKindNewBlock {
			cancel()
		} else {
			require.FailNow(t, "Invalid event was sent out")
		}
	}))

	block, _ := chain.Store().LatestBlock()
	pub.BlockApplied(chain.ChainID, block)

	<-ctx.Done()

	require.ErrorIs(t, ctx.Err(), context.Canceled)
}
