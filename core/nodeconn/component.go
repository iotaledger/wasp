package nodeconn

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/nodeconn"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:     "NodeConn",
			DepsFunc: func(cDeps dependencies) { deps = cDeps },
			Provide:  provide,
			Run:      run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
)

type dependencies struct {
	dig.In

	MetricsEnabled bool             `name:"metricsEnabled"`
	Metrics        *metrics.Metrics `optional:"true"`
	NodeConnection chain.NodeConnection
}

func provide(c *dig.Container) error {
	type nodeConnectionDeps struct {
		dig.In

		NodeBridge *nodebridge.NodeBridge
	}

	type nodeConnectionResult struct {
		dig.Out

		NodeConnection chain.NodeConnection
	}

	if err := c.Provide(func(deps nodeConnectionDeps) nodeConnectionResult {
		return nodeConnectionResult{
			NodeConnection: nodeconn.New(
				CoreComponent.Daemon().ContextStopped(),
				CoreComponent.Logger(),
				deps.NodeBridge,
			),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func run() error {
	if deps.MetricsEnabled {
		deps.NodeConnection.SetMetrics(deps.Metrics.GetNodeConnectionMetrics())
	}

	return nil
}
