// scmeta package runs integration tests by calling WebAPi to itself for SC meta data
package nodeping

import (
	"github.com/iotaledger/goshimmer/packages/waspconn"
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
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Disabled, configure, run)
	log    *logger.Logger
)

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

//orig1, _ := testplugins.CreateOriginData(testplugins.SC1, nil)
//orig2, _ := testplugins.CreateOriginData(testplugins.SC2, nil)
//orig3, _ := testplugins.CreateOriginData(testplugins.SC3, nil)
//log.Infof("origin transaction SC1 ID = %s", orig1.ID().String())
//log.Infof("origin transaction SC2 ID = %s", orig2.ID().String())
//log.Infof("origin transaction SC3 ID = %s", orig3.ID().String())
