package nodeconn

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/inx-app/nodebridge"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/plugins/metrics"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "NodeConn",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Configure: configure,
			Run:       run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies

	nc chain.NodeConnection
)

type dependencies struct {
	dig.In
	NodeBridge     *nodebridge.NodeBridge
	MetricsEnabled bool `name:"metricsEnabled"`
}

func configure() error {
	nc = nodeconn.New(
		CoreComponent.Daemon().ContextStopped(),
		CoreComponent.Logger(),
		deps.NodeBridge,
	)

	return nil
}

func run() error {
	if deps.MetricsEnabled {
		nc.SetMetrics(metrics.AllMetrics().GetNodeConnectionMetrics())
	}

	return nil
}

func NodeConnection() chain.NodeConnection {
	return nc
}
