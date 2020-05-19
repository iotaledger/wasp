package nodeconn

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/shutdown"
	"time"
)

// PluginName is the name of the database plugin.
const PluginName = "NodeConn"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		go nodeConnect()

		<-shutdownSignal
		log.Info("Stopping node connection..")
		go func() {
			bconnMutex.Lock()
			defer bconnMutex.Unlock()

			if bconn != nil {
				log.Infof("Closing connection with node..")
				_ = bconn.Close()
				log.Infof("Closing connection with node.. Done")
			}
		}()

		go func() {
			for {
				select {
				case <-shutdownSignal:
					return

				case <-time.After(1 * time.Second):
					sendSubscriptionsIfNeeded()
				}
			}
		}()
	}, shutdown.PriorityNodeConnection)
	if err != nil {
		log.Errorf("failed to start NodeConn worker")
	}
}
