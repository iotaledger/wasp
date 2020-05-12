package peering

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/shutdown"
)

// PluginName is the name of the database plugin.
const PluginName = "Peering"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	if err := daemon.BackgroundWorker("WaspPeering", func(shutdownSignal <-chan struct{}) {

		go connectOutboundLoop()
		go connectInboundLoop()

		<-shutdownSignal

		log.Info("Closing all connections with peers...")
		closeAll()
		log.Info("Closing all connections with peers... done")
	}, shutdown.PriorityPeering); err != nil {
		panic(err)
	}
}
