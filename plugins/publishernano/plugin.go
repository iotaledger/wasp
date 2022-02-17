package publishernano

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/publisher"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pub"
	_ "go.nanomsg.org/mangos/v3/transport/all"
)

// PluginName is the name of the Publisher plugin.
const PluginName = "PublisherNano"

var log *logger.Logger

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, nil, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	messages := make(chan []byte, 100)

	port := parameters.GetInt(parameters.NanomsgPublisherPort)
	socket, err := openSocket(port)
	if err != nil {
		log.Errorf("failed to initialize publisher: %v", err)
		return
	}
	log.Infof("nanomsg publisher is running on port %d", port)

	err = daemon.BackgroundWorker(PluginName, func(ctx context.Context) {
		for {
			select {
			case msg := <-messages:
				if socket != nil {
					err := socket.Send(msg)
					if err != nil {
						log.Errorf("Failed to publish message: %v", err)
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
	})
	if err != nil {
		panic(err)
	}

	publisher.Event.Attach(events.NewClosure(func(msgType string, parts []string) {
		msg := msgType + " " + strings.Join(parts, " ")
		select {
		case messages <- []byte(msg):
		case <-time.After(1 * time.Second):
			log.Warnf("Failed to publish message: [%s]", msg)
		}
	}))
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
