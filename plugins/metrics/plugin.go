package metrics

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
)

const PluginName = "Metrics"

var (
	log        *logger.Logger
	allMetrics *metrics.Metrics
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Disabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
	allMetrics = metrics.New(log)
}

func run(_ *node.Plugin) {
	if !parameters.GetBool(parameters.MetricsEnabled) {
		return
	}

	log.Infof("Starting %s ...", PluginName)
	if err := daemon.BackgroundWorker("Prometheus exporter", func(shutdownSignal <-chan struct{}) {
		log.Info("Starting Prometheus exporter ... done")

		bindAddr := parameters.GetString(parameters.MetricsBindAddress)
		stopped := make(chan struct{})
		go func() {
			defer close(stopped)
			allMetrics.Start(bindAddr)
		}()

		select {
		case <-shutdownSignal:
		case <-stopped:
		}
		log.Infof("Stopping %s ...", PluginName)
		defer log.Infof("Stopping %s ... done", PluginName)
		if err := allMetrics.Stop(); err != nil {
			log.Errorf("error stopping: %s", err)
		}
	}, parameters.PriorityMetrics); err != nil {
		log.Warnf("error starting as daemon: %s", err)
	}
}

func AllMetrics() *metrics.Metrics {
	return allMetrics
}
