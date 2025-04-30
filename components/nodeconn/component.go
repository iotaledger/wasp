package nodeconn

import (
	"context"
	"fmt"

	"go.uber.org/dig"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/shutdown"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/components/chains"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/nodeconn"
)

func init() {
	Component = &app.Component{
		Name:      "NodeConn",
		DepsFunc:  func(cDeps dependencies) { deps = cDeps },
		Params:    params,
		Provide:   provide,
		Configure: configure,
	}
}

var (
	Component *app.Component
	deps      dependencies
)

type dependencies struct {
	dig.In

	NodeConnection  chain.NodeConnection
	ShutdownHandler *shutdown.ShutdownHandler
}

func provide(c *dig.Container) error {
	type nodeConnectionDeps struct {
		dig.In

		ShutdownHandler *shutdown.ShutdownHandler
	}

	if err := c.Provide(func(deps nodeConnectionDeps) chain.NodeConnection {
		address := lo.Must(iotago.AddressFromHex(ParamsL1.PackageID))

		nodeConnection, err := nodeconn.New(
			Component.Daemon().ContextStopped(),
			*address,
			chains.ParamsChains.MempoolMaxOnledgerInPool,
			ParamsL1.WebsocketURL,
			ParamsL1.HTTPURL,
			Component.Logger.NewChildLogger("nc"),
			deps.ShutdownHandler,
		)
		if err != nil {
			Component.LogPanicf("Creating NodeConnection failed: %s", err.Error())
		}
		return nodeConnection
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}

func configure() error {
	if err := Component.Daemon().BackgroundWorker(Component.Name, func(ctx context.Context) {
		Component.LogInfof("Starting %s ... done", Component.Name)
		if err := deps.NodeConnection.Run(ctx); err != nil {
			deps.ShutdownHandler.SelfShutdown(fmt.Sprintf("Starting %s failed, error: %s", Component.Name, err.Error()), true)
		}
		Component.LogInfof("Stopping %s ... done", Component.Name)
	}, daemon.PriorityNodeConnection); err != nil {
		Component.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
