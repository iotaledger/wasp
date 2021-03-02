package ipfs

import (
	"context"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/ipfs"
)

// PluginName is the name of the web API plugin.
const PluginName = "IPFS"

var (
	log *logger.Logger // Logger for IPFS functionality
)

func Init() *node.Plugin {
	var configure, run func(*node.Plugin)
	configure = func(*node.Plugin) {
		log = logger.NewLogger(PluginName)
	}
	run = func(*node.Plugin) {
		var worker func(<-chan struct{}) = func(shutdownSignal <-chan struct{}) {
			var ctx context.Context
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(context.Background())
			defer func() {
				log.Infof("Stopping %s ...", PluginName)
				cancel()
				log.Infof("Stopping %s ... done", PluginName)
			}()

			var err error
			err = ipfs.Start(log, &ctx)
			if err != nil {
				log.Errorf("Error starting %s: %s", PluginName, err)
				return
			}
			log.Infof("%s started", PluginName)

			// stop if we are shutting down or the server could not be started
			select {
			case <-shutdownSignal:
			}

			// stopping will be done by cancel function from context
		}

		log.Infof("Starting %s ...", PluginName)
		var err error = daemon.BackgroundWorker("Ipfs node", worker, 5 /*parameters.PriorityIpfs*/)
		if err != nil {
			log.Errorf("Error starting as daemon: %s", err)
		}
	}

	Plugin := node.NewPlugin(PluginName, node.Enabled, configure, run)
	return Plugin
}
