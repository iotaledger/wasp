package metrics

import (
	"context"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
)

func init() {
	Plugin = &app.Plugin{
		Component: &app.Component{
			Name:           "Metrics",
			DepsFunc:       func(cDeps dependencies) { deps = cDeps },
			Params:         params,
			InitConfigPars: initConfigPars,
			Provide:        provide,
			Run:            run,
		},
		IsEnabled: func() bool {
			return ParamsMetrics.Enabled
		},
	}
}

var (
	Plugin *app.Plugin
	deps   dependencies
)

type dependencies struct {
	dig.In

	Metrics *metrics.Metrics `optional:"true"`
}

func initConfigPars(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		MetricsEnabled bool `name:"metricsEnabled"`
	}

	if err := c.Provide(func() cfgResult {
		return cfgResult{
			MetricsEnabled: ParamsMetrics.Enabled,
		}
	}); err != nil {
		Plugin.LogPanic(err)
	}

	return nil
}

func provide(c *dig.Container) error {
	if !ParamsMetrics.Enabled {
		return nil
	}

	type metricsResult struct {
		dig.Out

		Metrics *metrics.Metrics
	}

	if err := c.Provide(func() metricsResult {
		return metricsResult{
			Metrics: metrics.New(Plugin.Logger()),
		}
	}); err != nil {
		Plugin.LogPanic(err)
	}

	return nil
}

func run() error {
	if !ParamsMetrics.Enabled {
		return nil
	}

	Plugin.LogInfof("Starting %s ...", Plugin.Name)
	if err := Plugin.Daemon().BackgroundWorker("Prometheus exporter", func(ctx context.Context) {
		Plugin.LogInfo("Starting Prometheus exporter ... done")

		stopped := make(chan struct{})
		go func() {
			defer close(stopped)
			deps.Metrics.Start(ParamsMetrics.BindAddress)
		}()

		select {
		case <-ctx.Done():
		case <-stopped:
		}
		Plugin.LogInfof("Stopping %s ...", Plugin.Name)
		defer Plugin.LogInfof("Stopping %s ... done", Plugin.Name)

		if err := deps.Metrics.Stop(); err != nil { //nolint:contextcheck // false positive
			Plugin.LogErrorf("error stopping: %s", err)
		}
	}, parameters.PriorityMetrics); err != nil {
		Plugin.LogWarnf("error starting as daemon: %s", err)
	}

	return nil
}
