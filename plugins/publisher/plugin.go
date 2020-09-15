package publisher

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pub"
	_ "go.nanomsg.org/mangos/v3/transport/all"
)

// PluginName is the name of the Publisher plugin.
const PluginName = "Publisher"

var (
	log      *logger.Logger
	socket   mangos.Socket
	messages = make(chan []byte, 100)
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	port := parameters.GetInt(parameters.NanomsgPublisherPort)
	if err := openSocket(port); err != nil {
		log.Errorf("failed to initialize publisher: %v", err)
	} else {
		log.Infof("nanomsg publisher is running on port %d", port)
	}

	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		for {
			select {
			case msg := <-messages:
				if socket != nil {
					err := socket.Send(msg)
					if err != nil {
						log.Errorf("Failed to publish message: %v", err)
					}
				}
			case <-shutdownSignal:
				if socket != nil {
					socket.Close()
					socket = nil
				}
				return
			}
		}
	})
	if err != nil {
		panic(err)
	}
}

func openSocket(port int) error {
	var err error
	socket, err = pub.NewSocket()
	if err != nil {
		socket = nil
		return err
	}

	url := fmt.Sprintf("tcp://:%d", port)
	if err = socket.Listen(url); err != nil {
		socket = nil
		return err
	}
	return nil
}

func Publish(msgType string, parts ...string) {
	msg := msgType
	for _, s := range parts {
		msg = msg + " " + s
	}
	select {
	case messages <- []byte(msg):
	case <-time.After(1 * time.Second):
		log.Warnf("Failed to publish message: [%s]", msg)
	}
}
