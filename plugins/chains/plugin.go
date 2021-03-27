package chains

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/chains"
)

const PluginName = "Chains"

var (
	log       *logger.Logger
	allChains *chains.Chains
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	log.Infof("running %s plugin..", PluginName)
	allChains = chains.New(log)
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		if err := allChains.ActivateAllFromRegistry(); err != nil {
			log.Errorf("failed to read chain activation records from registry: %v", err)
			return
		}
		allChains.Attach()

		<-shutdownSignal

		log.Info("dismissing chains...")
		go func() {
			allChains.Dismiss()
			log.Info("dismissing chains... Done")
		}()
	})
	if err != nil {
		log.Error(err)
		return
	}
}
