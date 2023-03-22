package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	websocketserver "nhooyr.io/websocket"

	"github.com/iotaledger/hive.go/app/configuration"
	appLogger "github.com/iotaledger/hive.go/app/logger"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/web/websockethub"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/solo"
)

func InitWebsocket(ctx context.Context, t *testing.T, eventsToSubscribe []publisher.ISCEventType) (*Service, *websockethub.Hub, *solo.Chain) {
	_ = appLogger.InitGlobalLogger(configuration.New())
	log := logger.NewLogger("Test")

	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Log: log}) //nolint:contextcheck

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
		websocketHub.Run(ctx)
	}()

	return ws, websocketHub, chain
}

func TestWebsocketEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	ws, _, chain := InitWebsocket(ctx, t, []publisher.ISCEventType{publisher.ISCEventKindNewBlock})

	// publisherEvent is the event that gets called from the websocket event handlers.
	//
	//		publisher -> TEvent(ISCEvent[T]) -> Websocket Service EventHandlers
	//			-> Mapping -> publisherEvent(ISCEvent) -> Websocket Send
	//
	// It's the last step before the events get send via the websocket to the client.
	// It's also the last step to validate the events without actually connecting with a websocket client.
	ws.publisherEvent.Hook(func(iscEvent *ISCEvent) {
		require.Exactly(t, iscEvent.ChainID, chain.ChainID.String())

		if iscEvent.Kind == publisher.ISCEventKindNewBlock {
			cancel()
		} else {
			require.FailNow(t, "Invalid event was sent out")
		}
	})

	// provoke a new block to be created
	chain.DepositBaseTokensToL2(1, nil)

	<-ctx.Done()

	require.ErrorIs(t, ctx.Err(), context.Canceled)
}
