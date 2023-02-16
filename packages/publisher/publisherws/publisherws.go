// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package publisherws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/subscriptionmanager"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/publisher"
)

type PublisherWebSocket struct {
	hub                 *websockethub.Hub
	log                 *logger.Logger
	msgTypes            map[string]bool
	subscriptionManager *subscriptionmanager.SubscriptionManager[websockethub.ClientID, string]
}

func New(log *logger.Logger, hub *websockethub.Hub, msgTypes []string) *PublisherWebSocket {
	msgTypesMap := make(map[string]bool)
	for _, t := range msgTypes {
		msgTypesMap[t] = true
	}

	subscriptionManager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[websockethub.ClientID, string](5),
	)

	return &PublisherWebSocket{
		hub:                 hub,
		log:                 log.Named("PublisherWebSocket"),
		msgTypes:            msgTypesMap,
		subscriptionManager: subscriptionManager,
	}
}

func (p *PublisherWebSocket) hasSubscribedToAllChains(session *websockethub.Client) bool {
	return p.subscriptionManager.ClientSubscribedToTopic(session.ID(), "chains")
}

func (p *PublisherWebSocket) hasSubscribedToSingleChain(session *websockethub.Client, chainID isc.ChainID) bool {
	return p.subscriptionManager.ClientSubscribedToTopic(session.ID(), fmt.Sprintf("chains/%s", chainID.String()))
}

func (p *PublisherWebSocket) createEventWriter(ctx context.Context, session *websockethub.Client) *events.Closure {
	eventClosure := events.NewClosure(func(event *publisher.ISCEvent) {
		if event == nil {
			return
		}

		if !p.msgTypes[event.Kind] {
			return
		}

		if !p.hasSubscribedToAllChains(session) && !p.hasSubscribedToSingleChain(session, event.ChainID) {
			return
		}

		if !p.subscriptionManager.ClientSubscribedToTopic(session.ID(), event.Kind) {
			return
		}

		session.Send(ctx, event)
	})

	return eventClosure
}

type BaseCommand struct {
	Command string `json:"command"`
}

const (
	CommandSubscribe           = "subscribe"
	CommandClientWasSubscribed = "client_subscribed"

	CommandUnsubscribe           = "unsubscribe"
	CommandClientWasUnsubscribed = "client_unsubscribed"
)

type SubscriptionCommand struct {
	Command string `json:"command"`
	Topic   string `json:"topic"`
}

func (p *PublisherWebSocket) handleSubscriptionCommand(ctx context.Context, client *websockethub.Client, message []byte) {
	p.log.Info(string(message))

	var command SubscriptionCommand
	if err := json.Unmarshal(message, &command); err != nil {
		p.log.Warnf("Could not deserialize message to type ControlCommand, msg: '%v'", message)
		return
	}

	var err error

	switch command.Command {
	case CommandSubscribe:
		p.subscriptionManager.Subscribe(client.ID(), command.Topic)
		err = client.Send(ctx, SubscriptionCommand{
			Command: CommandClientWasSubscribed,
			Topic:   command.Topic,
		})
	case CommandUnsubscribe:
		p.subscriptionManager.Unsubscribe(client.ID(), command.Topic)
		err = client.Send(ctx, SubscriptionCommand{
			Command: CommandClientWasUnsubscribed,
			Topic:   command.Topic,
		})
	}

	if err != nil {
		p.log.Warnf("Failed to send response: %v %v %v", err, ctx.Err(), ctx)
	}
}

func (p *PublisherWebSocket) handleNodeCommands(ctx context.Context, client *websockethub.Client, message []byte) {
	var baseCommand BaseCommand
	if err := json.Unmarshal(message, &baseCommand); err != nil {
		p.log.Warnf("Could not deserialize message to type BaseCommand")
		return
	}

	switch baseCommand.Command {
	case CommandSubscribe, CommandUnsubscribe:
		p.handleSubscriptionCommand(ctx, client, message)
	default:
		p.log.Warnf("Could not deserialize message")
	}
}

func (p *PublisherWebSocket) OnClientCreated(ctx context.Context, client *websockethub.Client) {
	client.ReceiveChan = make(chan *websockethub.WebsocketMsg, 100)

	eventWriter := p.createEventWriter(ctx, client)
	publisher.Event.Hook(eventWriter)
	defer publisher.Event.Detach(eventWriter)

	go func() {
		for {
			select {
			case <-client.ExitSignal:
				// client was disconnected
				return

			case msg, ok := <-client.ReceiveChan:
				if !ok {
					// client was disconnected
					return
				}

				p.handleNodeCommands(ctx, client, msg.Data)
			}
		}
	}()
}

func (p *PublisherWebSocket) OnConnect(client *websockethub.Client, request *http.Request) {
	p.log.Infof("accepted websocket connection from %s", request.RemoteAddr)
	p.subscriptionManager.Connect(client.ID())
}

func (p *PublisherWebSocket) OnDisconnect(client *websockethub.Client, request *http.Request) {
	p.subscriptionManager.Disconnect(client.ID())
	p.log.Infof("closed websocket connection from %s", request.RemoteAddr)
}

// ServeHTTP serves the websocket.
// Provide a chainID to filter for a certain chain, provide an empty chain id to get all chain events.
func (p *PublisherWebSocket) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) error {
	return p.hub.ServeWebsocket(responseWriter, request,
		func(client *websockethub.Client) {
			p.OnClientCreated(request.Context(), client)
		}, func(client *websockethub.Client) {
			p.OnConnect(client, request)
		}, func(client *websockethub.Client) {
			p.OnDisconnect(client, request)
		})
}
