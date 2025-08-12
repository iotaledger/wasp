// Package websocket implements the webapi websocket connection
package websocket

import (
	"github.com/iotaledger/hive.go/web/subscriptionmanager"
	"github.com/iotaledger/hive.go/web/websockethub"
	"github.com/iotaledger/wasp/v2/packages/publisher"
)

type SubscriptionValidator struct {
	messageTypes        map[publisher.ISCEventType]bool
	subscriptionManager *subscriptionmanager.SubscriptionManager[websockethub.ClientID, string]
}

func NewSubscriptionValidator(messageTypes map[publisher.ISCEventType]bool, subscriptionManager *subscriptionmanager.SubscriptionManager[websockethub.ClientID, string]) *SubscriptionValidator {
	return &SubscriptionValidator{
		messageTypes:        messageTypes,
		subscriptionManager: subscriptionManager,
	}
}

// shouldProcessEvent validates if any subscriber has subscribed to a certain chainID and messageType.
// it returns false if no one has subscribed to those parameters
// this usually means, that there is no need to process a certain incoming event.
func (p *SubscriptionValidator) shouldProcessEvent(messageType publisher.ISCEventType) bool {
	if !p.messageTypes[messageType] {
		return false
	}

	if !p.subscriptionManager.TopicHasSubscribers(string(messageType)) {
		return false
	}

	return true
}

// isClientAllowed validates if a certain subscriber has subscribed to a certain chainID and messageType.
// it returns false if the client has not subscribed to those parameters
// this usually means, that there is no need to process a certain outgoing event.
func (p *SubscriptionValidator) isClientAllowed(client *websockethub.Client, messageType publisher.ISCEventType) bool {
	if !p.messageTypes[messageType] {
		return false
	}

	if !p.subscriptionManager.ClientSubscribedToTopic(client.ID(), string(messageType)) {
		return false
	}

	return true
}
