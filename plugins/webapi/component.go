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
	"go.uber.org/zap"
	"nhooyr.io/websocket"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/hive.go/core/configuration"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/inx-app/pkg/httpserver"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/webapi"
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

const (
	broadcastQueueSize            = 20000
	clientSendChannelSize         = 1000
	webSocketWriteTimeout         = time.Duration(3) * time.Second
	maxWebsocketMessageSize int64 = 510
)

type dependencies struct {
	dig.In

	Echo         *echo.Echo          `name:"webapiEcho"`
	EchoSwagger  echoswagger.ApiRoot `name:"webapiServer"`
	WebsocketHub *websockethub.Hub   `name:"websocketHub"`
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

func CreateEchoSwagger(e *echo.Echo, version string) echoswagger.ApiRoot {
	echoSwagger := echoswagger.New(e, "/doc", &echoswagger.Info{
		Title:       "Wasp API",
		Description: "REST API for the Wasp node",
		Version:     version,
	})

	echoSwagger.AddSecurityAPIKey("Authorization", "JWT Token", echoswagger.SecurityInHeader).
		SetExternalDocs("Find out more about Wasp", "https://wiki.iota.org/smart-contracts/overview").
		SetUI(echoswagger.UISetting{DetachSpec: false, HideTop: false}).
		SetScheme("http", "https")

	echoSwagger.SetRequestContentType(echo.MIMEApplicationJSON)
	echoSwagger.SetResponseContentType(echo.MIMEApplicationJSON)

	return echoSwagger
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
		Publisher                   *publisher.Publisher
	}

	type webapiServerResult struct {
		dig.Out

		Echo         *echo.Echo          `name:"webapiEcho"`
		EchoSwagger  echoswagger.ApiRoot `name:"webapiServer"`
		WebsocketHub *websockethub.Hub   `name:"websocketHub"`
	}

	if err := c.Provide(func(deps webapiServerDeps) webapiServerResult {
		e := httpserver.NewEcho(
			Plugin.Logger(),
			nil,
			ParamsWebAPI.DebugRequestLoggerEnabled,
		)

		e.Server.ReadTimeout = ParamsWebAPI.Limits.ReadTimeout
		e.Server.WriteTimeout = ParamsWebAPI.Limits.WriteTimeout

		e.HidePort = true
		e.HTTPErrorHandler = webapi.CompatibilityHTTPErrorHandler(Plugin.Logger().WithOptions(zap.AddStacktrace(zap.ErrorLevel)))

		// timeout middleware
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				timeoutCtx, cancel := context.WithTimeout(c.Request().Context(), ParamsWebAPI.Limits.Timeout)
				defer cancel()

				c.SetRequest(c.Request().WithContext(timeoutCtx))

				return next(c)
			}
		})
		e.Use(middleware.BodyLimit(ParamsWebAPI.Limits.MaxBodyLength))
		e.Use(apmechov4.Middleware())

		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
		}))

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowHeaders:     []string{"*"},
			AllowMethods:     []string{"*"},
			AllowCredentials: true,
		}))

		echoSwagger := CreateEchoSwagger(e, deps.AppInfo.Version)
		websocketOptions := websocket.AcceptOptions{
			InsecureSkipVerify: true,
			// Disable compression due to incompatibilities with the latest Safari browsers:
			// https://github.com/tilt-dev/tilt/issues/4746
			CompressionMode: websocket.CompressionDisabled,
		}

		hub := websockethub.NewHub(Plugin.Logger(), &websocketOptions, broadcastQueueSize, clientSendChannelSize, maxWebsocketMessageSize)

		webapi.Init(
			Plugin.App().NewLogger("WebAPI/v2"),
			echoSwagger,
			hub,
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
			deps.Publisher,
		)

		return webapiServerResult{
			Echo:         e,
			EchoSwagger:  echoSwagger,
			WebsocketHub: hub,
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

			deps.Echo.Server.BaseContext = func(_ net.Listener) context.Context {
				// set BaseContext to be the same as the plugin, so that requests being processed don't hang the shutdown procedure
				return ctx
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

	if err := Plugin.Daemon().BackgroundWorker("WebsocketHub", func(ctx context.Context) {
		go func() {
			go deps.WebsocketHub.Run(ctx)
			<-ctx.Done()
			Plugin.LogInfo("Stopping WebAPI[WS]")
		}()
	}, daemon.PriorityWebAPI); err != nil {
		Plugin.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
