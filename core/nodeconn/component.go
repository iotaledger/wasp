package nodeconn

import (
	"context"
	"time"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/plugins/metrics"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "NodeConn",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Configure: configure,
			Run:       run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies

	nc          chain.NodeConnection
	initialized = ready.New("NodeConn")
)

type dependencies struct {
	dig.In

	MetricsEnabled bool `name:"metricsEnabled"`
}

func configure() error {
	nc = nodeconn.New(
		nodeconn.ChainL1Config{
			INXAddress: ParamsNodeconn.INXAddress,
		},
		CoreComponent.Logger(),
	)

	return nil
}

func run() error {
	err := CoreComponent.Daemon().BackgroundWorker(CoreComponent.Name, func(ctx context.Context) {
		if deps.MetricsEnabled {
			nc.SetMetrics(metrics.AllMetrics().GetNodeConnectionMetrics())
		}
		defer nc.Close()

		initialized.SetReady()

		<-ctx.Done()

		CoreComponent.LogInfo("Stopping node connection..")
	}, parameters.PriorityNodeConnection)
	if err != nil {
		CoreComponent.LogErrorf("failed to start NodeConn worker")
	}

	return err
}

func NodeConnection() chain.NodeConnection {
	initialized.MustWait(5 * time.Second)
	return nc
}
