package chains

import (
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	_ "github.com/iotaledger/wasp/packages/chain/chainimpl"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/processors"
	"github.com/iotaledger/wasp/plugins/registry"
)

const PluginName = "Chains"

var (
	log         *logger.Logger
	allChains   *chains.Chains
	initialized = ready.New(PluginName)
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	allChains = chains.New(
		log,
		processors.Config,
		parameters.GetInt(parameters.OffledgerBroadcastUpToNPeers),
		time.Duration(parameters.GetInt(parameters.OffledgerBroadcastInterval))*time.Millisecond,
		parameters.GetBool(parameters.PullMissingRequestsFromCommittee),
	)
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		if err := allChains.ActivateAllFromRegistry(registry.DefaultRegistry); err != nil {
			log.Errorf("failed to read chain activation records from registry: %v", err)
			return
		}
		allChains.Attach(nodeconn.NodeConnection())

		initialized.SetReady()

		<-shutdownSignal

		log.Info("dismissing chains...")
		go func() {
			allChains.Dismiss()
			log.Info("dismissing chains... Done")
		}()
	}, parameters.PriorityChains)
	if err != nil {
		log.Error(err)
		return
	}
}

func AllChains() *chains.Chains {
	initialized.MustWait(5 * time.Second)
	return allChains
}
