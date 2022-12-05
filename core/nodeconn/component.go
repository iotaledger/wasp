package nodeconn

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/nodeconn"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:    "NodeConn",
			Provide: provide,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
)

func provide(c *dig.Container) error {

	type nodeConnectionMetricsResult struct {
		dig.Out

		NodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics
	}

	if err := c.Provide(func() nodeConnectionMetricsResult {
		return nodeConnectionMetricsResult{
			NodeConnectionMetrics: nodeconnmetrics.New(),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	type nodeConnectionDeps struct {
		dig.In

		NodeBridge            *nodebridge.NodeBridge
		NodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics
	}

	type nodeConnectionResult struct {
		dig.Out

		NodeConnection chain.NodeConnection
	}

	if err := c.Provide(func(deps nodeConnectionDeps) nodeConnectionResult {
		return nodeConnectionResult{
			NodeConnection: nodeconn.New(
				CoreComponent.Daemon().ContextStopped(),
				CoreComponent.Logger().Named("nc"),
				deps.NodeBridge,
				deps.NodeConnectionMetrics,
			),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}
