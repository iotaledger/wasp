package commands

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"testing"
	"time"

	websocketserver "github.com/coder/websocket"
	"github.com/stretchr/testify/require"

	appLogger "github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/web/subscriptionmanager"
	"github.com/iotaledger/hive.go/web/websockethub"
)

func initTest() (*CommandManager, *websockethub.Hub, context.CancelFunc) {
	log := appLogger.NewLogger(appLogger.WithName("Test"))

	ctx, cancel := context.WithTimeout(context.Background(), testmisc.GetTimeout(5*time.Second))

	subscriptionManager := subscriptionmanager.New[websockethub.ClientID, string]()
	subscriptionManager.Connect(1)

	manager := NewCommandHandler(log, subscriptionManager)
	hub := websockethub.NewHub(log.NewChildLogger("Hub"), &websocketserver.AcceptOptions{InsecureSkipVerify: true}, 500, 500, 500)

	go func() { hub.Run(ctx) }()

	// Test needs to wait a little until the hub has taken up the supplied context
	time.Sleep(1 * time.Second)

	return manager, hub, cancel
}

func sendNodeCommand(manager *CommandManager, client *websockethub.Client, command any) error {
	var messageBytes []byte
	var err error

	if messageBytes, err = json.Marshal(command); err != nil {
		return err
	}

	return manager.HandleNodeCommands(client, messageBytes)
}

func TestSuccessfulSubscription(t *testing.T) {
	manager, hub, _ := initTest()

	client := websockethub.NewClient(hub, nil, func(client *websockethub.Client) {}, func(client *websockethub.Client) {})

	_ = sendNodeCommand(manager, client, SubscriptionCommand{
		BaseCommand: BaseCommand{
			Command: CommandSubscribe,
		},
		Topic: "TEST",
	})

	require.True(t, manager.subscriptionManager.TopicHasSubscribers("TEST"))
}

// TestSuccessfulUnsubscription subscribes, then unsubscribes
func TestSuccessfulUnsubscription(t *testing.T) {
	manager, hub, _ := initTest()

	client := websockethub.NewClient(hub, nil, func(client *websockethub.Client) {}, func(client *websockethub.Client) {})

	_ = sendNodeCommand(manager, client, SubscriptionCommand{
		BaseCommand: BaseCommand{
			Command: CommandSubscribe,
		},
		Topic: "TEST",
	})

	require.True(t, manager.subscriptionManager.TopicHasSubscribers("TEST"))

	_ = sendNodeCommand(manager, client, SubscriptionCommand{
		BaseCommand: BaseCommand{
			Command: CommandUnsubscribe,
		},
		Topic: "TEST",
	})

	require.False(t, manager.subscriptionManager.TopicHasSubscribers("TEST"))
}

// TestFailingSubscription validates the returned and handled error
// As we have established no actual websocket connection, the response should always fail.
// In this test we force the context to be canceled to ignore timeouts.
func TestFailingSubscriptionDueToFailedSend(t *testing.T) {
	manager, hub, cancel := initTest()

	client := websockethub.NewClient(hub, nil, func(client *websockethub.Client) {}, func(client *websockethub.Client) {})

	// Force a fake cancelation of the websocket hub
	cancel()

	err := sendNodeCommand(manager, client, SubscriptionCommand{
		BaseCommand: BaseCommand{
			Command: CommandSubscribe,
		},
		Topic: "TEST",
	})

	require.ErrorIs(t, errors.Unwrap(err), ErrFailedToSendMessage)
}

func TestFailingSubscriptionDueToInvalidTopic(t *testing.T) {
	manager, hub, _ := initTest()

	client := websockethub.NewClient(hub, nil, func(client *websockethub.Client) {}, func(client *websockethub.Client) {})
	err := sendNodeCommand(manager, client, SubscriptionCommand{
		BaseCommand: BaseCommand{
			Command: CommandSubscribe,
		},
	})
	require.ErrorIs(t, errors.Unwrap(err), ErrFailedToValidateCommand)
}

func TestFailingSubscriptionDueToInvalidCommand(t *testing.T) {
	manager, hub, _ := initTest()

	client := websockethub.NewClient(hub, nil, func(client *websockethub.Client) {}, func(client *websockethub.Client) {})
	err := sendNodeCommand(manager, client, SubscriptionCommand{})
	require.ErrorIs(t, errors.Unwrap(err), ErrFailedToValidateCommand)
}
