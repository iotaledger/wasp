// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/generics/options"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/subscriptionmanager"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/webapi/websocket/commands"
)

type Service struct {
	commandHandler        *commands.CommandManager
	eventHandler          *EventHandler
	hub                   *websockethub.Hub
	log                   *logger.Logger
	publisherEvent        *event.Event[*ISCEvent]
	subscriptionManager   *subscriptionmanager.SubscriptionManager[websockethub.ClientID, string]
	subscriptionValidator *SubscriptionValidator

	maxTopicSubscriptionsPerClient int
}

func WithMaxTopicSubscriptionsPerClient(maxTopicSubscriptionsPerClient int) options.Option[Service] {
	return func(d *Service) {
		d.maxTopicSubscriptionsPerClient = maxTopicSubscriptionsPerClient
	}
}

func NewWebsocketService(log *logger.Logger, hub *websockethub.Hub, msgTypes []publisher.ISCEventType, pub *publisher.Publisher, opts ...options.Option[Service]) *Service {
	serviceOptions := options.Apply(&Service{
		maxTopicSubscriptionsPerClient: 0,
	}, opts)

	msgTypesMap := make(map[publisher.ISCEventType]bool)
	for _, t := range msgTypes {
		msgTypesMap[t] = true
	}

	publishEvent := event.New[*ISCEvent]()

	subscriptionManager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[websockethub.ClientID, string](serviceOptions.maxTopicSubscriptionsPerClient),
	)

	subscriptionValidator := NewSubscriptionValidator(msgTypesMap, subscriptionManager)
	eventHandler := NewEventHandler(pub, publishEvent, subscriptionValidator)
	commandHandler := commands.NewCommandHandler(log, subscriptionManager)

	return &Service{
		log:                   log.Named("Websocket Service"),
		hub:                   hub,
		commandHandler:        commandHandler,
		eventHandler:          eventHandler,
		publisherEvent:        publishEvent,
		subscriptionManager:   subscriptionManager,
		subscriptionValidator: subscriptionValidator,
	}
}

func (p *Service) onClientCreated(client *websockethub.Client) {
	client.ReceiveChan = make(chan *websockethub.WebsocketMsg, 100)

	go func() {
		eventWriter := event.NewClosure(func(iscEvent *ISCEvent) {
			if !p.subscriptionValidator.isClientAllowed(client, iscEvent.ChainID, iscEvent.Kind) {
				return
			}

			if err := client.Send(client.Context(), iscEvent); err != nil {
				p.log.Warnf("error sending message to client:[%d], err:[%v]", client.ID(), err)
			}
		})

		p.publisherEvent.Hook(eventWriter)
		defer p.publisherEvent.Detach(eventWriter)

		for {
			// we need to nest the client.ReceiveChan into the default case because
			// the select cases are executed in random order if multiple
			// conditions are true at the time of entry in the select case.
			select {
			case <-client.ExitSignal:
				// client was disconnected
				return
			default:
				select {
				case <-client.ExitSignal:
					// client was disconnected
					return
				case msg, ok := <-client.ReceiveChan:
					if !ok {
						// client was disconnected
						return
					}

					// returned error is currently only used in tests, as the node command handler handles errors already.
					_ = p.commandHandler.HandleNodeCommands(client, msg.Data)
				}
			}
		}
	}()
}

func (p *Service) onConnect(client *websockethub.Client, request *http.Request) {
	p.log.Infof("accepted websocket connection for client:[%d], from:[%s]", client.ID(), request.RemoteAddr)
	p.subscriptionManager.Connect(client.ID())
}

func (p *Service) onDisconnect(client *websockethub.Client, request *http.Request) {
	p.subscriptionManager.Disconnect(client.ID())
	p.log.Infof("closed websocket connection for client:[%d], from:[%s]", client.ID(), request.RemoteAddr)
}

// ServeHTTP serves the websocket.
func (p *Service) ServeHTTP(c echo.Context) error {
	return p.hub.ServeWebsocket(c.Response(), c.Request(),
		func(client *websockethub.Client) {
			p.onClientCreated(client)
		}, func(client *websockethub.Client) {
			p.onConnect(client, c.Request())
		}, func(client *websockethub.Client) {
			p.onDisconnect(client, c.Request())
		})
}

func (p *Service) EventHandler() *EventHandler {
	return p.eventHandler
}
