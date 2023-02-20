package nodeconn

import (
	"context"
	"fmt"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/nodeconn"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "NodeConn",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Provide:   provide,
			Configure: configure,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
)

type dependencies struct {
	dig.In

	NodeConnection  chain.NodeConnection
	ShutdownHandler *shutdown.ShutdownHandler
}

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
		ShutdownHandler       *shutdown.ShutdownHandler
	}

	type nodeConnectionResult struct {
		dig.Out

		NodeConnection chain.NodeConnection
	}

	if err := c.Provide(func(deps nodeConnectionDeps) nodeConnectionResult {
		nodeConnection, err := nodeconn.New(
			CoreComponent.Daemon().ContextStopped(),
			CoreComponent.Logger().Named("nc"),
			deps.NodeBridge,
			deps.NodeConnectionMetrics,
			deps.ShutdownHandler,
		)
		if err != nil {
			CoreComponent.LogPanicf("Creating NodeConnection failed: %s", err.Error())
		}
		return nodeConnectionResult{
			NodeConnection: nodeConnection,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func configure() error {
	if err := CoreComponent.Daemon().BackgroundWorker(CoreComponent.Name, func(ctx context.Context) {
		CoreComponent.LogInfof("Starting %s ... done", CoreComponent.Name)
		if err := deps.NodeConnection.Run(ctx); err != nil {
			deps.ShutdownHandler.SelfShutdown(fmt.Sprintf("Starting %s failed, error: %s", CoreComponent.Name, err.Error()), true)
		}
		CoreComponent.LogInfof("Stopping %s ... done", CoreComponent.Name)
	}, daemon.PriorityNodeConnection); err != nil {
		CoreComponent.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
