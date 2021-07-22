package metrics

import (
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/ready"
)

const PluginName = "Metrics"

var (
	log         *logger.Logger
	allMetrics  *metrics.Metrics
	initialized = ready.New(PluginName)
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

		stopped := make(chan struct{})
		go func() {
			defer close(stopped)
			if err := allMetrics.Start(); err != nil {
				log.Warnf("Error serving: %s", err)
			}
		}()

		initialized.SetReady()
		select {
		case <-shutdownSignal:
		case <-stopped:
		}
		log.Info("Stopping %s ...", PluginName)
		defer log.Infof("Stopping %s ... done", PluginName)
		if err := allMetrics.Stop(); err != nil {
			log.Errorf("Error stopping: %s", err)
		}
	}, parameters.PriorityMetrics); err != nil {
		log.Warnf("Error starting as daemon: %s", err)
	}
}

func AllMetrics() *metrics.Metrics {
	initialized.MustWait(5 * time.Second)
	return allMetrics
}
