package nodeconn

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/plugins/peering"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "NodeConn"

const dialTimeout = 1 * time.Second

var (
	log *logger.Logger

	nc          chain.NodeConnection
	initialized = ready.New("NodeConn")
)

// Init initializes the plugin
func Init() *node.Plugin {
	return node.NewPlugin(PluginName, nil, node.Enabled, configure, run)
}

func NodeConnection() chain.NodeConnection {
	initialized.MustWait(5 * time.Second)
	return nc
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(ctx context.Context) {
		addr := parameters.GetString(parameters.NodeConnHost)
		port := parameters.GetInt(parameters.NodeConnPort)
		networkProvider := peering.DefaultNetworkProvider()
		nc = nodeconn.New(addr, port, networkProvider, log)
		defer nc.Close()

		initialized.SetReady()

		<-ctx.Done()

		log.Info("Stopping node connection..")
	}, parameters.PriorityNodeConnection)
	if err != nil {
		log.Errorf("failed to start NodeConn worker")
	}
}
