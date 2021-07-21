package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const PluginName = "Metrics"

var log *logger.Logger

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Disabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	if !parameters.GetBool(parameters.MetricsEnabled) {
		return
	}

	log.Infof("Starting %s ...", PluginName)
	if err := daemon.BackgroundWorker("Prometheus exporter", func(shutdownSignal <-chan struct{}) {
		log.Info("Starting Prometheus exporter ... done")

		e := echo.New()
		e.HideBanner = true
		e.Use(middleware.Recover())

		e.GET("/metrics", func(c echo.Context) error {
			handler := promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{
					EnableOpenMetrics: true,
				},
			)
			handler.ServeHTTP(c.Response(), c.Request())
			return nil
		})

		bindAddr := parameters.GetString(parameters.MetricsBindAddress)
		server := &http.Server{Addr: bindAddr, Handler: e}

		stopped := make(chan struct{})
		go func() {
			defer close(stopped)
			log.Infof("%s started, bind-address=%s", PluginName, bindAddr)
			if err := server.ListenAndServe(); err != nil {
				log.Warn("Error serving: %s", err)
			}
		}()
		select {
		case <-shutdownSignal:
		case <-stopped:
		}
		log.Info("Stopping %s ...", PluginName)
		defer log.Infof("Stopping %s ... done", PluginName)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Errorf("Error stopping: %s", err)
		}
	}, parameters.PriorityMetrics); err != nil {
		log.Warnf("Error starting as daemon: %s", err)
	}
}
