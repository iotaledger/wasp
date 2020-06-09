package publisher

import (
	"fmt"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/config"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pub"
	_ "go.nanomsg.org/mangos/v3/transport/all"
	"sync"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "Publisher"

var (
	Plugin    = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log       *logger.Logger
	socket    mangos.Socket
	sockMutex sync.RWMutex
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		port := config.Node.GetInt(CfgNanomsgPublisherPort)
		if err := openSocket(port); err != nil {
			log.Errorf("failed to initialize publisher: %v", err)
			return
		}
		log.Infof("nanomsg publisher is running on port %d", port)

		<-shutdownSignal

		sockMutex.Lock()

		socket.Close()
		socket = nil

		sockMutex.Unlock()

		log.Infof("publisher has been closed")
	})
	if err != nil {
		log.Error(err)
		return
	}
}

func openSocket(port int) error {
	sockMutex.Lock()
	defer sockMutex.Unlock()

	var err error
	socket, err = pub.NewSocket()
	if err != nil {
		socket = nil
		return err
	}
	url := fmt.Sprintf("tcp://:%d", port)
	log.Debugf("listening to %s", url)

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
	sockMutex.RLock()
	defer sockMutex.RUnlock()

	if socket != nil {
		_ = socket.Send([]byte(msg))
	}
}
