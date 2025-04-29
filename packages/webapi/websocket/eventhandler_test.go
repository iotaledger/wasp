package websocket

import (
	"context"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	appLogger "github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/hive.go/web/subscriptionmanager"
	"github.com/iotaledger/hive.go/web/websockethub"

	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func initTest(ctx context.Context) (*publisher.Publisher, *EventHandler, *event.Event1[*ISCEvent], *subscriptionmanager.SubscriptionManager[websockethub.ClientID, string]) {
	log := appLogger.NewLogger(appLogger.WithName("Test"))

	pub := publisher.New(log)

	go func() {
		pub.Run(ctx)
	}()

	publisherEvent := event.New1[*ISCEvent]()

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
	ctx, cancel := context.WithTimeout(context.Background(), testmisc.GetTimeout(5*time.Second))
	pub, _, publisherEvent, subscriptionManager := initTest(ctx)

	subscriptionManager.Subscribe(1, string(publisher.ISCEventKindNewBlock))

	chainID := isctest.RandomChainID()

	publisherEvent.Hook(func(iscEvent *ISCEvent) {
		require.Exactly(t, iscEvent.ChainID, chainID.String())
		cancel()
	})

	pub.Events.NewBlock.Trigger(&publisher.ISCEvent[*publisher.BlockWithTrieRoot]{
		Kind:    publisher.ISCEventKindNewBlock,
		ChainID: chainID,
		Issuer:  isctest.NewRandomAgentID(),
		Payload: &publisher.BlockWithTrieRoot{
			BlockInfo: &blocklog.BlockInfo{},
			TrieRoot:  trie.Hash{},
		},
	})

	<-ctx.Done()
	require.ErrorIs(t, ctx.Err(), context.Canceled, "The context was not correctly canceled by the event receiver and timed out. "+
		"This means no event was sent and this needs to fail the test")
}
