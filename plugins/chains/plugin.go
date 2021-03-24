package chains

import (
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/plugins/nodeconn"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
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
	if err := nodeconn.Ready.Wait(); err != nil {
		log.Errorf("failed waiting for NodeConn plugin ready. Abort %s plugin. Err = %v", PluginName, err)
		return
	}
	allChains = chains.New(log, nodeconn.NodeConn)
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		if err := allChains.ActivateAllFromRegistry(); err != nil {
			log.Errorf("failed to read chain activation records from registry: %v", err)
			return
		}
		allChains.Attach()

		<-shutdownSignal

		log.Info("dismissing chains...")
		go func() {
			allChains.Detach()
			allChains.Dismiss()
			log.Info("dismissing chains... Done")
		}()
	})
	if err != nil {
		log.Error(err)
		return
	}
}
