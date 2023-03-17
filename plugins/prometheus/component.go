package prometheus

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

// routeMetrics is the route for getting the prometheus metrics.
// GET returns metrics.
const (
	routeMetrics      = "/metrics"
	readHeaderTimeout = 500 * time.Millisecond
)

func init() {
	Plugin = &app.Plugin{
		Component: &app.Component{
			Name:      "Prometheus",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Provide:   provide,
			Configure: configure,
			Run:       run,
		},
		IsEnabled: func() bool {
			return ParamsPrometheus.Enabled
		},
	}
}

var (
	Plugin *app.Plugin
	deps   dependencies
)

type dependencies struct {
	dig.In

	PrometheusEcho     *echo.Echo `name:"prometheusEcho"`
	PrometheusRegistry *prometheus.Registry

	AppInfo               *app.Info
	NodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics
	WebAPIEcho            *echo.Echo                  `name:"webapiEcho" optional:"true"`
	BlockWALMetrics       *smGPAUtils.BlockWALMetrics `optional:"true"`
}

func provide(c *dig.Container) error {
	type depsOut struct {
		dig.Out
		PrometheusEcho     *echo.Echo `name:"prometheusEcho"`
		PrometheusRegistry *prometheus.Registry
	}

	if err := c.Provide(func() depsOut {
		e := echo.New()
		e.HideBanner = true
		e.Use(middleware.Recover())
		e.Server.ReadHeaderTimeout = readHeaderTimeout

		return depsOut{
			PrometheusEcho:     e,
			PrometheusRegistry: prometheus.NewRegistry(),
		}
	}); err != nil {
		Plugin.LogPanic(err)
	}

	return nil
}

func configure() error {
	if ParamsPrometheus.NodeMetrics {
		configureNode(deps.PrometheusRegistry, deps.AppInfo)
	}
	if ParamsPrometheus.NodeConnMetrics {
		deps.NodeConnectionMetrics.Register(deps.PrometheusRegistry)
	}
	if ParamsPrometheus.BlockWALMetrics && deps.BlockWALMetrics != nil {
		deps.BlockWALMetrics.Register(deps.PrometheusRegistry)
	}
	if ParamsPrometheus.RestAPIMetrics {
		configureRestAPI(deps.PrometheusRegistry, deps.WebAPIEcho)
	}
	if ParamsPrometheus.GoMetrics {
		deps.PrometheusRegistry.MustRegister(collectors.NewGoCollector())
	}
	if ParamsPrometheus.ProcessMetrics {
		deps.PrometheusRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	return nil
}

func run() error {
	Plugin.LogInfo("Starting Prometheus exporter ...")

	if err := Plugin.Daemon().BackgroundWorker("Prometheus exporter", func(ctx context.Context) {
		Plugin.LogInfo("Starting Prometheus exporter ... done")

		deps.PrometheusEcho.GET(routeMetrics, func(c echo.Context) error {
			handler := promhttp.HandlerFor(
				deps.PrometheusRegistry,
				promhttp.HandlerOpts{
					EnableOpenMetrics: true,
				},
			)

			if ParamsPrometheus.PromhttpMetrics {
				handler = promhttp.InstrumentMetricHandler(deps.PrometheusRegistry, handler)
			}

			handler.ServeHTTP(c.Response().Writer, c.Request())

			return nil
		})

		bindAddr := ParamsPrometheus.BindAddress

		go func() {
			Plugin.LogInfof("You can now access the Prometheus exporter using: http://%s/metrics", bindAddr)
			if err := deps.PrometheusEcho.Start(bindAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
				Plugin.LogWarnf("Stopped Prometheus exporter due to an error (%s)", err)
			}
		}()

		<-ctx.Done()
		Plugin.LogInfo("Stopping Prometheus exporter ...")

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCtxCancel()

		//nolint:contextcheck // false positive
		err := deps.PrometheusEcho.Shutdown(shutdownCtx)
		if err != nil {
			Plugin.LogWarn(err)
		}

		Plugin.LogInfo("Stopping Prometheus exporter ... done")
	}, daemon.PriorityPrometheus); err != nil {
		Plugin.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
