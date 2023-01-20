package webapi

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pangpanglabs/echoswagger/v2"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/hive.go/core/configuration"
	"github.com/iotaledger/inx-app/pkg/httpserver"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/webapi"
	v1 "github.com/iotaledger/wasp/packages/webapi/v1"
	v2 "github.com/iotaledger/wasp/packages/webapi/v2"
)

func init() {
	Plugin = &app.Plugin{
		Component: &app.Component{
			Name:           "WebAPI",
			DepsFunc:       func(cDeps dependencies) { deps = cDeps },
			Params:         params,
			InitConfigPars: initConfigPars,
			Provide:        provide,
			Run:            run,
		},
		IsEnabled: func() bool {
			return ParamsWebAPI.Enabled
		},
	}
}

var (
	Plugin *app.Plugin
	deps   dependencies
)

type dependencies struct {
	dig.In

	EchoSwagger echoswagger.ApiRoot `name:"webapiServer"`
}

func initConfigPars(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		WebAPIBindAddress string `name:"webAPIBindAddress"`
	}

	if err := c.Provide(func() cfgResult {
		return cfgResult{
			WebAPIBindAddress: ParamsWebAPI.BindAddress,
		}
	}); err != nil {
		Plugin.LogPanic(err)
	}

	return nil
}

//nolint:funlen
func provide(c *dig.Container) error {
	type webapiServerDeps struct {
		dig.In

		AppInfo                     *app.Info
		AppConfig                   *configuration.Configuration `name:"appConfig"`
		ShutdownHandler             *shutdown.ShutdownHandler
		APICacheTTL                 time.Duration `name:"apiCacheTTL"`
		PublisherPort               int           `name:"publisherPort"`
		Chains                      *chains.Chains
		NodeConnectionMetrics       nodeconnmetrics.NodeConnectionMetrics
		ChainRecordRegistryProvider registry.ChainRecordRegistryProvider
		DKShareRegistryProvider     registry.DKShareRegistryProvider
		NodeIdentityProvider        registry.NodeIdentityProvider
		NetworkProvider             peering.NetworkProvider       `name:"networkProvider"`
		TrustedNetworkManager       peering.TrustedNetworkManager `name:"trustedNetworkManager"`
		Node                        *dkg.Node
		UserManager                 *users.UserManager
	}

	type webapiServerResult struct {
		dig.Out

		Echo        *echo.Echo          `name:"webapiEcho"`
		EchoSwagger echoswagger.ApiRoot `name:"webapiServer"`
	}

	if err := c.Provide(func(deps webapiServerDeps) webapiServerResult {
		e := httpserver.NewEcho(
			Plugin.Logger(),
			nil,
			ParamsWebAPI.DebugRequestLoggerEnabled,
		)

		e.Server.ReadTimeout = ParamsWebAPI.ReadTimeout
		e.Server.WriteTimeout = ParamsWebAPI.WriteTimeout

		e.HidePort = true
		e.HTTPErrorHandler = webapi.CompatibilityHTTPErrorHandler(Plugin.Logger())

		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
		}))

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowHeaders:     []string{"*"},
			AllowMethods:     []string{"*"},
			AllowCredentials: true,
		}))

		// TODO using this middleware hides the stack trace https://github.com/golang/go/issues/27375
		e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			ErrorMessage: "request timeout exceeded",
			Timeout:      1 * time.Minute,
		}))

		echoSwagger := echoswagger.New(e, "/doc", &echoswagger.Info{
			Title:       "Wasp API",
			Description: "REST API for the Wasp node",
			Version:     deps.AppInfo.Version,
		})

		echoSwagger.AddSecurityAPIKey("Authorization", "JWT Token", echoswagger.SecurityInHeader).
			SetExternalDocs("Find out more about Wasp", "https://wiki.iota.org/smart-contracts/overview").
			SetUI(echoswagger.UISetting{DetachSpec: false, HideTop: false}).
			SetScheme("http", "https")

		echoSwagger.SetRequestContentType(echo.MIMEApplicationJSON)
		echoSwagger.SetResponseContentType(echo.MIMEApplicationJSON)

		echoSwagger.AddSecurityAPIKey("Authorization", "JWT Token", echoswagger.SecurityInHeader).
			SetExternalDocs("Find out more about Wasp", "https://wiki.iota.org/smart-contracts/overview").
			SetUI(echoswagger.UISetting{DetachSpec: false, HideTop: false}).
			SetScheme("http", "https")

		v1.Init(
			Plugin.App().NewLogger("WebAPI/v1"),
			echoSwagger,
			deps.AppInfo.Version,
			deps.NetworkProvider,
			deps.TrustedNetworkManager,
			deps.UserManager,
			deps.ChainRecordRegistryProvider,
			deps.DKShareRegistryProvider,
			deps.NodeIdentityProvider,
			func() *chains.Chains {
				return deps.Chains
			},
			func() *dkg.Node {
				return deps.Node
			},
			func() {
				deps.ShutdownHandler.SelfShutdown("wasp was shutdown via API", false)
			},
			deps.NodeConnectionMetrics,
			ParamsWebAPI.Auth,
			ParamsWebAPI.NodeOwnerAddresses,
			deps.APICacheTTL,
			deps.PublisherPort,
		)

		v2.Init(
			Plugin.App().NewLogger("WebAPI/v2"),
			echoSwagger,
			deps.AppInfo.Version,
			deps.AppConfig,
			deps.NetworkProvider,
			deps.TrustedNetworkManager,
			deps.UserManager,
			deps.ChainRecordRegistryProvider,
			deps.DKShareRegistryProvider,
			deps.NodeIdentityProvider,
			func() *chains.Chains {
				return deps.Chains
			},
			func() *dkg.Node {
				return deps.Node
			},
			deps.ShutdownHandler,
			deps.NodeConnectionMetrics,
			ParamsWebAPI.Auth,
			ParamsWebAPI.NodeOwnerAddresses,
			deps.APICacheTTL,
		)

		return webapiServerResult{
			Echo:        e,
			EchoSwagger: echoSwagger,
		}
	}); err != nil {
		Plugin.LogPanic(err)
	}

	return nil
}

func run() error {
	Plugin.LogInfof("Starting %s server ...", Plugin.Name)
	if err := Plugin.Daemon().BackgroundWorker(Plugin.Name, func(ctx context.Context) {
		Plugin.LogInfof("Starting %s server ... done", Plugin.Name)

		go func() {
			Plugin.LogInfof("You can now access the WebAPI using: http://%s", ParamsWebAPI.BindAddress)
			if err := deps.EchoSwagger.Echo().Start(ParamsWebAPI.BindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
				Plugin.LogWarnf("Stopped %s server due to an error (%s)", Plugin.Name, err)
			}
			deps.EchoSwagger.Echo().Server.BaseContext = func(_ net.Listener) context.Context {
				// set BaseContext to be the same as the plugin, so that requests being processed don't hang the shutdown procedure
				return ctx
			}
		}()

		<-ctx.Done()
		Plugin.LogInfof("Stopping %s server ...", Plugin.Name)

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCtxCancel()

		//nolint:contextcheck // false positive
		if err := deps.EchoSwagger.Echo().Shutdown(shutdownCtx); err != nil {
			Plugin.LogWarn(err)
		}

		Plugin.LogInfof("Stopping %s server ... done", Plugin.Name)
	}, daemon.PriorityWebAPI); err != nil {
		Plugin.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
