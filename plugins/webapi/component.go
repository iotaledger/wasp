package webapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pangpanglabs/echoswagger/v2"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/core/shutdown"
	"github.com/iotaledger/wasp/core/chains"
	"github.com/iotaledger/wasp/core/dkg"
	"github.com/iotaledger/wasp/core/peering"
	"github.com/iotaledger/wasp/core/registry"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/wasp"
	"github.com/iotaledger/wasp/packages/webapi"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/plugins/metrics"
)

func init() {
	Plugin = &app.Plugin{
		Component: &app.Component{
			Name:     "WebAPI",
			DepsFunc: func(cDeps dependencies) { deps = cDeps },
			Params:   params,
			Run:      run,
		},
		IsEnabled: func() bool {
			return true
		},
	}
}

var (
	Plugin *app.Plugin
	deps   dependencies

	Server     echoswagger.ApiRoot
	allMetrics *metricspkg.Metrics
)

type dependencies struct {
	dig.In

	ShutdownHandler *shutdown.ShutdownHandler
	MetricsEnabled  bool `name:"metricsEnabled"`
	WAL             *wal.WAL
	APICacheTTL     time.Duration `name:"apiCacheTTL"`
	PublisherPort   int           `name:"publisherPort"`
}

func run() error {
	Plugin.LogInfof("Starting %s ...", Plugin.Name)
	if err := Plugin.Daemon().BackgroundWorker("WebAPI Server", worker, parameters.PriorityWebAPI); err != nil {
		Plugin.LogErrorf("error starting as daemon: %s", err)
	}

	return nil
}

func worker(ctx context.Context) {
	initWebAPI()
	stopped := make(chan struct{})
	server := Server.Echo()
	go func() {
		defer close(stopped)
		bindAddr := ParamsWebAPI.BindAddress
		Plugin.LogInfof("%s started, bind-address=%s", Plugin.Name, bindAddr)
		if err := server.Start(bindAddr); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				Plugin.LogErrorf("error serving: %s", err)
			}
		}
	}()

	// stop if we are shutting down or the server could not be started
	select {
	case <-ctx.Done():
	case <-stopped:
	}

	Plugin.LogInfof("Stopping %s ...", Plugin.Name)
	defer Plugin.LogInfof("Stopping %s ... done", Plugin.Name)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		Plugin.LogErrorf("error stopping: %s", err)
	}
}

func initWebAPI() {
	Server = echoswagger.New(echo.New(), "/doc", &echoswagger.Info{
		Title:       "Wasp API",
		Description: "REST API for the Wasp node",
		Version:     wasp.Version,
	})

	Server.Echo().HideBanner = true
	Server.Echo().HidePort = true
	Server.Echo().HTTPErrorHandler = httperrors.HTTPErrorHandler
	Server.Echo().Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))
	Server.Echo().Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{"*"},
	}))

	network := peering.DefaultNetworkProvider()
	if network == nil {
		panic("dependency NetworkProvider is missing in WebAPI")
	}
	tnm := peering.DefaultTrustedNetworkManager()
	if tnm == nil {
		panic("dependency TrustedNetworkManager is missing in WebAPI")
	}
	if deps.MetricsEnabled {
		allMetrics = metrics.AllMetrics()
	}
	webapi.Init(
		Server,
		network,
		tnm,
		registry.DefaultRegistry,
		chains.AllChains,
		dkg.DefaultNode,
		func() {
			deps.ShutdownHandler.SelfShutdown("wasp was shutdown via API", false)
		},
		allMetrics,
		deps.WAL,
		ParamsWebAPI.Auth,
		ParamsWebAPI.NodeOwnerAddresses,
		deps.APICacheTTL,
		deps.PublisherPort,
	)
}
