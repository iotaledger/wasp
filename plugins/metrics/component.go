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
			Params:         params,
			InitConfigPars: initConfigPars,
			Configure:      configure,
			Run:            run,
		},
		IsEnabled: func() bool {
			return ParamsMetrics.Enabled
		},
	}
}

var (
	Plugin     *app.Plugin
	allMetrics *metrics.Metrics
)

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

func configure() error {
	allMetrics = metrics.New(Plugin.Logger())

	return nil
}

func run() error {
	Plugin.LogInfof("Starting %s ...", Plugin.Name)
	if err := Plugin.Daemon().BackgroundWorker("Prometheus exporter", func(ctx context.Context) {
		Plugin.LogInfo("Starting Prometheus exporter ... done")

		stopped := make(chan struct{})
		go func() {
			defer close(stopped)
			allMetrics.Start(ParamsMetrics.BindAddress)
		}()

		select {
		case <-ctx.Done():
		case <-stopped:
		}
		Plugin.LogInfof("Stopping %s ...", Plugin.Name)
		defer Plugin.LogInfof("Stopping %s ... done", Plugin.Name)
		if err := allMetrics.Stop(); err != nil {
			Plugin.LogErrorf("error stopping: %s", err)
		}
	}, parameters.PriorityMetrics); err != nil {
		Plugin.LogWarnf("error starting as daemon: %s", err)
	}

	return nil
}

func AllMetrics() *metrics.Metrics {
	return allMetrics
}
