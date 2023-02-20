package inx

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/wasp/packages/daemon"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "INX",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Provide:   provide,
			Configure: configure,
		},
	}
}

type dependencies struct {
	dig.In
	NodeBridge      *nodebridge.NodeBridge
	ShutdownHandler *shutdown.ShutdownHandler
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
)

func provide(c *dig.Container) error {
	return c.Provide(func() (*nodebridge.NodeBridge, error) {
		nodeBridge := nodebridge.NewNodeBridge(
			CoreComponent.Logger(),
			nodebridge.WithTargetNetworkName(ParamsINX.TargetNetworkName),
		)

		if err := nodeBridge.Connect(
			CoreComponent.Daemon().ContextStopped(),
			ParamsINX.Address,
			ParamsINX.MaxConnectionAttempts,
		); err != nil {
			return nil, err
		}

		return nodeBridge, nil
	})
}

func configure() error {
	return CoreComponent.Daemon().BackgroundWorker("INX", func(ctx context.Context) {
		CoreComponent.LogInfo("Starting NodeBridge ...")
		deps.NodeBridge.Run(ctx)
		CoreComponent.LogInfo("Stopped NodeBridge")

		if !errors.Is(ctx.Err(), context.Canceled) {
			deps.ShutdownHandler.SelfShutdown("INX connection to node dropped", true)
		}
	}, daemon.PriorityDisconnectINX)
}
