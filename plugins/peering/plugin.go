package peering

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	"go.uber.org/atomic"
)

// PluginName is the name of the database plugin.
const PluginName = "Peering"

var (
	log         *logger.Logger
	initialized atomic.Bool
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
	if err := checkMyNetworkID(); err != nil {
		log.Errorf("can't continue: %v", err)
		return
	}
	log.Infof("my network Id = %s", MyNetworkId())
	initialized.Store(true)
}

func run(_ *node.Plugin) {
	if !initialized.Load() {
		return
	}
	if err := daemon.BackgroundWorker("WaspPeering", func(shutdownSignal <-chan struct{}) {

		go connectOutboundLoop()
		go connectInboundLoop()

		<-shutdownSignal

		log.Info("Closing all connections with peers...")
		closeAll()
		log.Info("Closing all connections with peers... done")
	}, parameters.PriorityPeering); err != nil {
		panic(err)
	}
}
