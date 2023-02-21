package websocket

import (
	"context"
	"encoding/json"

	"github.com/iotaledger/hive.go/core/websockethub"
)

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

func (p *Publisher) handleSubscriptionCommand(ctx context.Context, client *websockethub.Client, message []byte) {
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
		p.log.Warnf("error sending message: %v", err)
	}
}

func (p *Publisher) handleNodeCommands(client *websockethub.Client, message []byte) {
	var baseCommand BaseCommand
	if err := json.Unmarshal(message, &baseCommand); err != nil {
		p.log.Warnf("Could not deserialize message to type BaseCommand")
		return
	}

	switch baseCommand.Command {
	case CommandSubscribe, CommandUnsubscribe:
		p.handleSubscriptionCommand(client.Context(), client, message)
	default:
		p.log.Warnf("Could not deserialize message")
	}
}
