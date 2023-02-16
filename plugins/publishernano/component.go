package publishernano

import (
	"context"
	"encoding/json"
	"fmt"

	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pub"
	_ "go.nanomsg.org/mangos/v3/transport/all"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/publisher"
)

func init() {
	Plugin = &app.Plugin{
		Component: &app.Component{
			Name:           "PublisherNano",
			DepsFunc:       func(cDeps dependencies) { deps = cDeps },
			Params:         params,
			InitConfigPars: initConfigPars,
			Run:            run,
		},
		IsEnabled: func() bool {
			return ParamsPublisher.Enabled
		},
	}
}

var (
	Plugin *app.Plugin
	deps   dependencies
)

type dependencies struct {
	dig.In

	Publisher *publisher.Publisher
}

func initConfigPars(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		PublisherPort int `name:"publisherPort"`
	}

	if err := c.Provide(func() cfgResult {
		return cfgResult{
			PublisherPort: ParamsPublisher.Port,
		}
	}); err != nil {
		Plugin.LogPanic(err)
	}

	return nil
}

func run() error {
	messages := make(chan []byte, 100)

	port := ParamsPublisher.Port
	socket, err := openSocket(port)
	if err != nil {
		Plugin.LogErrorf("failed to initialize publisher: %w", err)
		return err
	}
	Plugin.LogInfof("nanomsg publisher is running on port %d", port)

	err = Plugin.Daemon().BackgroundWorker(Plugin.Name, func(ctx context.Context) {
		for {
			select {
			case msg := <-messages:
				if socket != nil {
					err2 := socket.Send(msg)
					if err2 != nil {
						Plugin.LogErrorf("failed to publish message: %w", err2)
					}
				}
			case <-ctx.Done():
				if socket != nil {
					_ = socket.Close()
					socket = nil
				}
				return
			}
		}
	}, daemon.PriorityNanoMsg)
	if err != nil {
		panic(err)
	}

	deps.Publisher.Events.Published.Hook(event.NewClosure(func(ev *publisher.ISCEvent) {
		msg, err := json.Marshal(ev)

		if err != nil {
			Plugin.LogWarnf("Could not marshal ISCEvent %s", ev.Kind)
			return
		}

		select {
		case messages <- msg:
		default:
			Plugin.LogWarnf("Failed to publish message: [%s]", msg)
		}
	}))

	return nil
}

func openSocket(port int) (mangos.Socket, error) {
	socket, err := pub.NewSocket()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("tcp://:%d", port)
	if err := socket.Listen(url); err != nil {
		return nil, err
	}
	return socket, nil
}
