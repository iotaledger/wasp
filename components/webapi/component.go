// Package webapi implements the REST API server component for the Wasp node.
package webapi

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	websocketserver "github.com/coder/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pangpanglabs/echoswagger/v2"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/hive.go/app/shutdown"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/web/websockethub"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/webapi"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/httpserver"
	"github.com/iotaledger/wasp/packages/webapi/websocket"
)

func init() {
	Component = &app.Component{
		Name:             "WebAPI",
		DepsFunc:         func(cDeps dependencies) { deps = cDeps },
		Params:           params,
		InitConfigParams: initConfigParams,
		IsEnabled:        func(_ *dig.Container) bool { return ParamsWebAPI.Enabled },
		Provide:          provide,
		Run:              run,
	}
}

var (
	Component *app.Component
	deps      dependencies
)

const (
	broadcastQueueSize            = 20000
	clientSendChannelSize         = 1000
	maxWebsocketMessageSize int64 = 510
)

type dependencies struct {
	dig.In

	EchoSwagger        echoswagger.ApiRoot `name:"webapiServer"`
	WebsocketHub       *websockethub.Hub   `name:"websocketHub"`
	NodeConnection     chain.NodeConnection
	WebsocketPublisher *websocket.Service `name:"websocketService"`
}

func initConfigParams(c *dig.Container) error {
	type cfgResult struct {
		dig.Out
		WebAPIBindAddress string `name:"webAPIBindAddress"`
	}

	if err := c.Provide(func() cfgResult {
		return cfgResult{
			WebAPIBindAddress: ParamsWebAPI.BindAddress,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}

func NewEcho(params *ParametersWebAPI, metrics *metrics.ChainMetricsProvider, log log.Logger) *echo.Echo {
	e := httpserver.NewEcho(
		log,
		nil,
		ParamsWebAPI.DebugRequestLoggerEnabled,
	)

	e.HideBanner = true

	e.Server.ReadTimeout = params.Limits.ReadTimeout
	e.Server.WriteTimeout = params.Limits.WriteTimeout

	e.HidePort = true
	e.HTTPErrorHandler = apierrors.HTTPErrorHandler()

	webapi.ConfirmedStateLagThreshold = params.Limits.ConfirmedStateLagThreshold
	authentication.DefaultJWTDuration = params.Auth.JWTConfig.Duration

	e.Pre(middleware.RemoveTrailingSlash())

	// publish metrics to prometheus component (that exposes a separate http server on another port)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Path(), "/chains/") {
				// ignore metrics for all requests not related to "chains/<chainID>""
				return next(c)
			}
			start := time.Now()
			err := next(c)

			status := c.Response().Status
			if err != nil {
				return err
			}

			chainID, ok := c.Get(controllerutils.EchoContextKeyChainID).(isc.ChainID)
			if !ok {
				return nil
			}

			operation, ok := c.Get(controllerutils.EchoContextKeyOperation).(string)
			if !ok {
				return nil
			}
			metrics.GetChainMetrics(chainID).WebAPI.WebAPIRequest(operation, status, time.Since(start))
			return nil
		}
	})

	// timeout middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			timeoutCtx, cancel := context.WithTimeout(c.Request().Context(), params.Limits.Timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(timeoutCtx))

			return next(c)
		}
	})

	// Middleware to unescape any supplied path (/path/foo%40bar/) parameter
	// Query parameters (?name=foo%40bar) get unescaped by default.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			escapedPathParams := c.ParamValues()
			unescapedPathParams := make([]string, len(escapedPathParams))

			for i, param := range escapedPathParams {
				unescapedParam, err := url.PathUnescape(param)

				if err != nil {
					unescapedPathParams[i] = param
				} else {
					unescapedPathParams[i] = unescapedParam
				}
			}

			c.SetParamValues(unescapedPathParams...)

			return next(c)
		}
	})

	e.Use(middleware.BodyLimit(params.Limits.MaxBodyLength))

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowCredentials: true,
	}))

	return e
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
		Chains                      *chains.Chains
		ChainMetricsProvider        *metrics.ChainMetricsProvider
		ChainRecordRegistryProvider registry.ChainRecordRegistryProvider
		DKShareRegistryProvider     registry.DKShareRegistryProvider
		NodeIdentityProvider        registry.NodeIdentityProvider
		NetworkProvider             peering.NetworkProvider       `name:"networkProvider"`
		TrustedNetworkManager       peering.TrustedNetworkManager `name:"trustedNetworkManager"`
		Node                        *dkg.Node
		UserManager                 *users.UserManager
		Publisher                   *publisher.Publisher
		NodeConn                    chain.NodeConnection
	}

	type webapiServerResult struct {
		dig.Out

		Echo               *echo.Echo          `name:"webapiEcho"`
		EchoSwagger        echoswagger.ApiRoot `name:"webapiServer"`
		WebsocketHub       *websockethub.Hub   `name:"websocketHub"`
		WebsocketPublisher *websocket.Service  `name:"websocketService"`
	}

	if err := c.Provide(func(deps webapiServerDeps) webapiServerResult {
		e := NewEcho(ParamsWebAPI, deps.ChainMetricsProvider, Component.Logger)

		echoSwagger := CreateEchoSwagger(e, deps.AppInfo.Version)
		websocketOptions := websocketserver.AcceptOptions{
			InsecureSkipVerify: true,
			// Disable compression due to incompatibilities with the latest Safari browsers:
			// https://github.com/tilt-dev/tilt/issues/4746
			CompressionMode: websocketserver.CompressionDisabled,
		}

		logger := Component.App().NewChildLogger("WebAPI/v2")

		hub := websockethub.NewHub(Component.Logger, &websocketOptions, broadcastQueueSize, clientSendChannelSize, maxWebsocketMessageSize)

		websocketService := websocket.NewWebsocketService(logger, hub, []publisher.ISCEventType{
			publisher.ISCEventKindNewBlock,
			publisher.ISCEventKindReceipt,
			publisher.ISCEventIssuerVM,
			publisher.ISCEventKindBlockEvents,
		}, deps.Publisher, websocket.WithMaxTopicSubscriptionsPerClient(ParamsWebAPI.Limits.MaxTopicSubscriptionsPerClient))

		if ParamsWebAPI.DebugRequestLoggerEnabled {
			echoSwagger.Echo().Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
				logger.LogDebugf("API Dump: Request=%q, Response=%q", reqBody, resBody)
			}))
		}

		webapi.Init(
			logger,
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
			deps.ChainMetricsProvider,
			ParamsWebAPI.Auth,
			deps.APICacheTTL,
			websocketService,
			ParamsWebAPI.IndexDBPath,
			ParamsWebAPI.AccountDumpsPath,
			deps.Publisher,
			deps.NodeConn.L1ParamsFetcher(),
			deps.NodeConn.L1Client(),
			jsonrpc.NewParameters(
				ParamsWebAPI.Limits.Jsonrpc.MaxBlocksInLogsFilterRange,
				ParamsWebAPI.Limits.Jsonrpc.MaxLogsInResult,
				ParamsWebAPI.Limits.Jsonrpc.WebsocketRateLimitMessagesPerSecond,
				ParamsWebAPI.Limits.Jsonrpc.WebsocketRateLimitBurst,
				ParamsWebAPI.Limits.Jsonrpc.WebsocketConnectionCleanupDuration,
				ParamsWebAPI.Limits.Jsonrpc.WebsocketClientBlockDuration,
				ParamsWebAPI.Limits.Jsonrpc.WebsocketRateLimitEnabled,
			),
		)

		return webapiServerResult{
			EchoSwagger:        echoSwagger,
			WebsocketHub:       hub,
			WebsocketPublisher: websocketService,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}

func run() error {
	Component.LogInfof("Starting %s server ...", Component.Name)
	if err := Component.Daemon().BackgroundWorker(Component.Name, func(ctx context.Context) {
		Component.LogInfof("Starting %s server ...", Component.Name)
		if err := deps.NodeConnection.WaitUntilInitiallySynced(ctx); err != nil {
			Component.LogErrorf("failed to start %s, waiting for L1 node to become sync failed, error: %s", err.Error())
			return
		}

		Component.LogInfof("Starting %s server ... done", Component.Name)

		go func() {
			deps.EchoSwagger.Echo().Server.BaseContext = func(_ net.Listener) context.Context {
				// set BaseContext to be the same as the plugin, so that requests being processed don't hang the shutdown procedure
				return ctx
			}

			Component.LogInfof("You can now access the WebAPI using: http://%s", ParamsWebAPI.BindAddress)
			if err := deps.EchoSwagger.Echo().Start(ParamsWebAPI.BindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
				Component.LogWarnf("Stopped %s server due to an error (%s)", Component.Name, err)
			}
		}()

		<-ctx.Done()

		Component.LogInfof("Stopping %s server ...", Component.Name)

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCtxCancel()

		//nolint:contextcheck // false positive
		if err := deps.EchoSwagger.Echo().Shutdown(shutdownCtx); err != nil {
			Component.LogWarn(err.Error())
		}

		Component.LogInfof("Stopping %s server ... done", Component.Name)
	}, daemon.PriorityWebAPI); err != nil {
		Component.LogPanicf("failed to start worker: %s", err)
	}

	if err := Component.Daemon().BackgroundWorker("WebAPI[WS]", func(ctx context.Context) {
		unhook := deps.WebsocketPublisher.EventHandler().AttachToEvents()
		defer unhook()

		deps.WebsocketHub.Run(ctx)
		Component.LogInfo("Stopping WebAPI[WS]")
	}, daemon.PriorityWebAPI); err != nil {
		Component.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
