package prometheus

import (
	"net/http"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const PluginName = "Prometheus"

var (
	log *logger.Logger
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	log.Info("Starting Prometheus exporter ...")

	if err := daemon.BackgroundWorker("Prometheus exporter", func(shutdownSignal <-chan struct{}) {
		log.Info("Starting Prometheus exporter ... done")

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
		<-shutdownSignal
	}, shutdown.PriotiyPrometheus); err != nil {
		log.Panic(err)
	}
}
