// scmeta package runs integration tests by calling WebAPi to itself for SC meta data
package roundtrip

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/committees"
	"time"
)

// PluginName is the name of the database plugin.
const PluginName = "TestingRoundTrip"

var (
	log *logger.Logger
)

const scAddress = "exZup69X1XwRNHiWWjoYy75aPNgC22YKkPV7sUJSBYA9"

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Disabled, configure, run)
}

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
				if c := committees.CommitteeByAddress(addr); c != nil {
					c.InitTestRound()
				}
			}
		}
	})
	if err != nil {
		log.Errorf("can't start worker: %v", err)
	}
}
