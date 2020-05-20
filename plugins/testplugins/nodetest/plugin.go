// scmeta package runs integration tests by calling WebAPi to itself for SC meta data
package nodetest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/dispatcher"
	"github.com/iotaledger/wasp/plugins/testplugins"
	"time"
)

// PluginName is the name of the database plugin.
const PluginName = "TestingNodeServices"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, testplugins.Status(PluginName), configure, run)
	log    *logger.Logger
)

const scAddress = "exZup69X1XwRNHiWWjoYy75aPNgC22YKkPV7sUJSBYA9"

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	addr, err := address.FromBase58(scAddress)
	if err != nil {
		log.Errorf("wrong testing data: %v", err)
		return
	}
	err = daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		for {
			select {
			case <-shutdownSignal:
				return
			case <-time.After(5 * time.Second):
				if c := dispatcher.CommitteeByAddress(&addr); c != nil {
					c.InitTestRound()
				}
			}
		}
	})
	if err != nil {
		log.Errorf("can't start worker: %v", err)
	}
}
