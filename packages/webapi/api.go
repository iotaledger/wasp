package webapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/hive.go/app/shutdown"
	loggerpkg "github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	userspkg "github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/webapi/controllers/chain"
	"github.com/iotaledger/wasp/packages/webapi/controllers/corecontracts"
	apimetrics "github.com/iotaledger/wasp/packages/webapi/controllers/metrics"
	"github.com/iotaledger/wasp/packages/webapi/controllers/node"
	"github.com/iotaledger/wasp/packages/webapi/controllers/requests"
	"github.com/iotaledger/wasp/packages/webapi/controllers/users"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/services"
	"github.com/iotaledger/wasp/packages/webapi/websocket"
)

const APIVersion = 1

func addHealthEndpoint(server echoswagger.ApiRoot) {
	server.GET("/health", func(c echo.Context) error { return c.NoContent(http.StatusOK) }).
		AddResponse(http.StatusOK, "The node is healthy.", nil, nil).
		SetOperationId("getHealth").
		SetSummary("Returns 200 if the node is healthy.")
}

func loadControllers(server echoswagger.ApiRoot, mocker *Mocker, controllersToLoad []interfaces.APIController, authMiddleware func() echo.MiddlewareFunc) {
	for _, controller := range controllersToLoad {
		group := server.Group(controller.Name(), fmt.Sprintf("/v%d/", APIVersion))
		controller.RegisterPublic(group, mocker)

		adminGroup := &APIGroupModifier{
			group: group,
			OverrideHandler: func(api echoswagger.Api) {
				// Force each route to set the security rule 'Authorization'
				api.SetSecurity("Authorization")

				// Any route in this group can fail due to invalid authorization
				api.AddResponse(http.StatusUnauthorized,
					"Unauthorized (Wrong permissions, missing token)", authentication.ValidationError{}, nil)
			},
		}

		if authMiddleware != nil {
			group.EchoGroup().Use(authMiddleware())
		}

		controller.RegisterAdmin(adminGroup, mocker)
	}
}

func Init(
	logger *loggerpkg.Logger,
	server echoswagger.ApiRoot,
	waspVersion string,
	config *configuration.Configuration,
	networkProvider peering.NetworkProvider,
	trustedNetworkManager peering.TrustedNetworkManager,
	userManager *userspkg.UserManager,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	chainsProvider chains.Provider,
	dkgNodeProvider dkg.NodeProvider,
	shutdownHandler *shutdown.ShutdownHandler,
	chainMetricsProvider *metrics.ChainMetricsProvider,
	authConfig authentication.AuthConfiguration,
	nodeOwnerAddresses []string,
	requestCacheTTL time.Duration,
	websocketService *websocket.Service,
	pub *publisher.Publisher,
	debugRequestLoggerEnabled bool,
) {
	// load mock files to generate correct echo swagger documentation
	mocker := NewMocker()
	mocker.LoadMockFiles()

	vmService := services.NewVMService(chainsProvider, chainRecordRegistryProvider)
	chainService := services.NewChainService(logger, chainsProvider, chainMetricsProvider, chainRecordRegistryProvider, vmService)
	committeeService := services.NewCommitteeService(chainsProvider, networkProvider, dkShareRegistryProvider)
	registryService := services.NewRegistryService(chainsProvider, chainRecordRegistryProvider)
	offLedgerService := services.NewOffLedgerService(chainService, networkProvider, requestCacheTTL)
	metricsService := services.NewMetricsService(chainsProvider, chainMetricsProvider)
	peeringService := services.NewPeeringService(chainsProvider, networkProvider, trustedNetworkManager)
	evmService := services.NewEVMService(chainService, networkProvider, pub, chainMetricsProvider, logger.Named("EVMService"))
	nodeService := services.NewNodeService(chainRecordRegistryProvider, nodeOwnerAddresses, nodeIdentityProvider, shutdownHandler, trustedNetworkManager)
	dkgService := services.NewDKGService(dkShareRegistryProvider, dkgNodeProvider, trustedNetworkManager)
	userService := services.NewUserService(userManager)
	// --

	claimValidator := func(claims *authentication.WaspClaims) bool {
		// The v2 api uses another way of permission handling, so we can always return true here.
		// Permissions are now validated at the route level. See the webapi/v2/controllers/*/controller.go routes.
		return true
	}

	authMiddleware := authentication.AddV2Authentication(server, userManager, nodeIdentityProvider, authConfig, claimValidator)

	controllersToLoad := []interfaces.APIController{
		chain.NewChainController(logger, chainService, committeeService, evmService, nodeService, offLedgerService, registryService, vmService),
		apimetrics.NewMetricsController(chainService, metricsService),
		node.NewNodeController(waspVersion, config, dkgService, nodeService, peeringService),
		requests.NewRequestsController(chainService, offLedgerService, peeringService, vmService),
		users.NewUsersController(userService),
		corecontracts.NewCoreContractsController(vmService),
	}

	if debugRequestLoggerEnabled {
		server.Echo().Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
			logger.Debugf("API Dump: Request=%q, Response=%q", reqBody, resBody)
		}))
	}

	addHealthEndpoint(server)
	addWebSocketEndpoint(server, websocketService)
	loadControllers(server, mocker, controllersToLoad, authMiddleware)
}
