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
	"github.com/iotaledger/wasp/plugins/metrics"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "NodeConn"

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
	nc = nodeconn.New(
		nodeconn.ChainL1Config{
			INXAddress: parameters.GetString(parameters.L1INXAddress),
		},
		log,
	)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(ctx context.Context) {
		if parameters.GetBool(parameters.MetricsEnabled) {
			nc.SetMetrics(metrics.AllMetrics().GetNodeConnectionMetrics())
		}
		defer nc.Close()

		initialized.SetReady()

		<-ctx.Done()

		log.Info("Stopping node connection..")
	}, parameters.PriorityNodeConnection)
	if err != nil {
		log.Errorf("failed to start NodeConn worker")
	}
}
