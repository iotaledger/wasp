// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/subscriptionmanager"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/publisher"
)

type Publisher struct {
	hub                 *websockethub.Hub
	log                 *logger.Logger
	msgTypes            map[string]bool
	subscriptionManager *subscriptionmanager.SubscriptionManager[websockethub.ClientID, string]
	publisher           *publisher.Publisher
}

func NewPublisher(log *logger.Logger, hub *websockethub.Hub, msgTypes []string, publisher *publisher.Publisher) *Publisher {
	msgTypesMap := make(map[string]bool)
	for _, t := range msgTypes {
		msgTypesMap[t] = true
	}

	subscriptionManager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[websockethub.ClientID, string](5),
	)

	return &Publisher{
		hub:                 hub,
		log:                 log.Named("Publisher"),
		msgTypes:            msgTypesMap,
		subscriptionManager: subscriptionManager,
		publisher:           publisher,
	}
}

func (p *Publisher) OnClientCreated(client *websockethub.Client) {
	client.ReceiveChan = make(chan *websockethub.WebsocketMsg, 100)

	eventWriter := p.createEventWriter(client.Context(), client)
	p.publisher.Events.Published.Hook(eventWriter)
	defer p.publisher.Events.Published.Detach(eventWriter)

	p.publisher.Events.BlockApplied.Hook()

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

			p.handleNodeCommands(client, msg.Data)
		}
	}
}

func (p *Publisher) OnConnect(client *websockethub.Client, request *http.Request) {
	p.log.Infof("accepted websocket connection from %s", request.RemoteAddr)
	p.subscriptionManager.Connect(client.ID())
}

func (p *Publisher) OnDisconnect(client *websockethub.Client, request *http.Request) {
	p.subscriptionManager.Disconnect(client.ID())
	p.log.Infof("closed websocket connection from %s", request.RemoteAddr)
}

// ServeHTTP serves the websocket.
// Provide a chainID to filter for a certain chain, provide an empty chain id to get all chain events.
func (p *Publisher) ServeHTTP(c echo.Context) error {
	p.log.Infof("ServeHTTP ctx: %v %v", c.Request().Context(), c.Request().Context().Err())

	return p.hub.ServeWebsocket(c.Response(), c.Request(),
		func(client *websockethub.Client) {
			go p.OnClientCreated(client)
		}, func(client *websockethub.Client) {
			p.OnConnect(client, c.Request())
		}, func(client *websockethub.Client) {
			p.OnDisconnect(client, c.Request())
		})
}
