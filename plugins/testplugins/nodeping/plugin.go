// scmeta package runs integration tests by calling WebAPi to itself for SC meta data
package nodeping

import (
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"math/rand"
	"time"
)

// PluginName is the name of the database plugin.
const PluginName = "TestingNodePing"

var (
	log *logger.Logger
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Disabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

const pingFrequency = 3 * time.Second

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		chCancel := make(chan struct{})
		go func() {
			for {
				select {
				case <-shutdownSignal:
					close(chCancel)
					return
				case <-chCancel:
					return
				case <-time.After(pingFrequency):
				}
				msgData, err := waspconn.EncodeMsg(&waspconn.WaspPingMsg{
					Id:        uint32(rand.Int()),
					Timestamp: time.Now().UnixNano(),
				})
				if err != nil {
					close(chCancel)
					return
				}
				err = nodeconn.SendDataToNode(msgData)
				if err != nil {
					log.Warnf("failed to send PING to node: %v", err)
				}
			}
		}()
		select {
		case <-shutdownSignal:
		case <-chCancel:
		}
	})
	if err != nil {
		log.Error(err)
	} else {
		log.Debug("started Node Ping plugin")
	}
}
