package commands

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/web/subscriptionmanager"
	"github.com/iotaledger/hive.go/web/websockethub"
)

const (
	CommandSubscribe   CommandType = "subscribe"
	CommandUnsubscribe CommandType = "unsubscribe"
)

type SubscriptionCommand struct {
	BaseCommand
	Topic string `json:"topic"`
}

const (
	EventClientWasSubscribed   EventType = "subscribed"
	EventClientWasUnsubscribed EventType = "unsubscribed"
)

type SubscriptionEvent struct {
	BaseEvent
	Topic string `json:"topic"`
}

type SubscriptionCommandHandler struct {
	log                 *logger.Logger
	subscriptionManager *subscriptionmanager.SubscriptionManager[websockethub.ClientID, string]
}

func (s *SubscriptionCommandHandler) SupportsCommand(commandType CommandType) bool {
	return commandType == CommandSubscribe || commandType == CommandUnsubscribe
}

func (s *SubscriptionCommandHandler) HandleCommand(client *websockethub.Client, message []byte) error {
	var command SubscriptionCommand
	var err error

	if err = json.Unmarshal(message, &command); err != nil {
		return errors.Wrap(ErrFailedToDeserializeCommand, err.Error())
	}

	if command.Topic == "" {
		return errors.Wrap(ErrFailedToValidateCommand, "Topic is empty")
	}

	switch command.Command {
	case CommandSubscribe:
		s.subscriptionManager.Subscribe(client.ID(), command.Topic)
		err = client.Send(client.Context(), SubscriptionEvent{
			BaseEvent: BaseEvent{
				Event: EventClientWasSubscribed,
			},
			Topic: command.Topic,
		})

	case CommandUnsubscribe:
		s.subscriptionManager.Unsubscribe(client.ID(), command.Topic)
		err = client.Send(client.Context(), SubscriptionEvent{
			BaseEvent: BaseEvent{
				Event: EventClientWasUnsubscribed,
			},
			Topic: command.Topic,
		})
	}

	if err != nil {
		return errors.Wrap(ErrFailedToSendMessage, err.Error())
	}

	return nil
}
