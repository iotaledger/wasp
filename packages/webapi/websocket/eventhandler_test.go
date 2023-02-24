package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/configuration"
	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/subscriptionmanager"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func initTest(ctx context.Context) (*publisher.Publisher, *EventHandler, *event.Event[*ISCEvent], *subscriptionmanager.SubscriptionManager[websockethub.ClientID, string]) {
	logger.InitGlobalLogger(configuration.New())
	log := logger.NewLogger("Test")

	pub := publisher.New(log)

	go func() {
		pub.Run(ctx)
	}()

	publisherEvent := event.New[*ISCEvent]()

	subscriptionManager := subscriptionmanager.New[websockethub.ClientID, string]()
	subscriptionManager.Connect(1)
	subscriptionManager.Subscribe(1, "chains")

	subscriptionValidator := NewSubscriptionValidator(map[publisher.ISCEventType]bool{
		publisher.ISCEventKindNewBlock: true,
	}, subscriptionManager)

	eventHandler := NewEventHandler(pub, publisherEvent, subscriptionValidator)
	eventHandler.AttachToEvents()

	return pub, eventHandler, publisherEvent, subscriptionManager
}

func TestSuccessfulEventHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	pub, _, publisherEvent, subscriptionManager := initTest(ctx)

	subscriptionManager.Subscribe(1, string(publisher.ISCEventKindNewBlock))

	chainID := isc.RandomChainID()

	publisherEvent.Attach(event.NewClosure(func(iscEvent *ISCEvent) {
		require.Exactly(t, iscEvent.ChainID, chainID.String())
		cancel()
	}))

	pub.Events.NewBlock.Trigger(&publisher.ISCEvent[*blocklog.BlockInfo]{
		Kind:    publisher.ISCEventKindNewBlock,
		ChainID: chainID,
		Issuer:  isc.NewRandomAgentID(),
		Payload: &blocklog.BlockInfo{},
	})

	<-ctx.Done()
	require.ErrorIs(t, ctx.Err(), context.Canceled, "The context was not correctly canceled by the event receiver and timed out. "+
		"This means no event was sent and this needs to fail the test")
}
