package websocket

import (
	"fmt"

	"github.com/iotaledger/hive.go/web/subscriptionmanager"
	"github.com/iotaledger/hive.go/web/websockethub"
	"github.com/iotaledger/wasp/packages/publisher"
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

func (p *SubscriptionValidator) hasClientSubscribedToAllChains(client *websockethub.Client) bool {
	return p.subscriptionManager.ClientSubscribedToTopic(client.ID(), "chains")
}

func (p *SubscriptionValidator) hasAnyoneSubscribedToAllChains() bool {
	return p.subscriptionManager.TopicHasSubscribers("chains")
}

func (p *SubscriptionValidator) hasClientSubscribedToSingleChain(client *websockethub.Client, chainID string) bool {
	return p.subscriptionManager.ClientSubscribedToTopic(client.ID(), fmt.Sprintf("chains/%s", chainID))
}

func (p *SubscriptionValidator) hasAnyoneSubscribedToSingleChain(chainID string) bool {
	return p.subscriptionManager.TopicHasSubscribers(fmt.Sprintf("chains/%s", chainID))
}

// shouldProcessEvent validates if any subscriber has subscribed to a certain chainID and messageType.
// it returns false if no one has subscribed to those parameters
// this usually means, that there is no need to process a certain incoming event.
func (p *SubscriptionValidator) shouldProcessEvent(chainID string, messageType publisher.ISCEventType) bool {
	if !p.messageTypes[messageType] {
		return false
	}

	// Check if any client has either subscribed to all chains [chains], or the supplied single chain id [chains/<chainID>]
	if !p.hasAnyoneSubscribedToAllChains() && !p.hasAnyoneSubscribedToSingleChain(chainID) {
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
func (p *SubscriptionValidator) isClientAllowed(client *websockethub.Client, chainID string, messageType publisher.ISCEventType) bool {
	if !p.messageTypes[messageType] {
		return false
	}

	// Check a client has either subscribed to all chains [chains], or the supplied single chain id [chains/<chainID>]
	if !p.hasClientSubscribedToAllChains(client) && !p.hasClientSubscribedToSingleChain(client, chainID) {
		return false
	}

	if !p.subscriptionManager.ClientSubscribedToTopic(client.ID(), string(messageType)) {
		return false
	}

	return true
}
