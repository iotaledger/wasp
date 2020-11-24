package peering

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	"go.uber.org/atomic"
)

const (
	pluginName = "Peering"
)

var (
	log         *logger.Logger
	initialized atomic.Bool
)

// Init is an entry point for this plugin.
func Init() *node.Plugin {
	return node.NewPlugin(pluginName, node.Enabled, configure, run)
}

// DefaultNetworkProvider returns the default network provider implementation.
func DefaultNetworkProvider() NetworkProvider {
	return nil // TODO
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(pluginName)
	if err := checkMyNetworkID(); err != nil {
		// can't continue because netid parameter is not correct
		log.Panicf("checkMyNetworkID: '%v'. || Check the 'netid' parameter in config.json", err)
		return
	}
	log.Infof("--------------------------------- netid is %s -----------------------------------", MyNetworkId())
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
