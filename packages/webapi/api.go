package webapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/hive.go/app/shutdown"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
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

var ConfirmedStateLagThreshold uint32

func AddHealthEndpoint(server echoswagger.ApiRoot, chainService interfaces.ChainService, metricsService interfaces.MetricsService) {
	server.GET("/health", func(e echo.Context) error {
		lag := metricsService.GetMaxChainConfirmedStateLag()
		if lag > ConfirmedStateLagThreshold {
			return e.String(http.StatusInternalServerError, fmt.Sprintf("chain unsync with %d diff", lag))
		}

		return e.NoContent(http.StatusOK)
	}).
		AddResponse(http.StatusOK, "The node is healthy.", nil, nil).
		SetOperationId("getHealth").
		SetSummary("Returns 200 if the node is healthy.")
}

func loadControllers(server echoswagger.ApiRoot, mocker *Mocker, controllersToLoad []interfaces.APIController, authMiddleware echo.MiddlewareFunc) {
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
			group.EchoGroup().Use(authMiddleware)
		}

		controller.RegisterAdmin(adminGroup, mocker)
	}
}

func Init(
	logger log.Logger,
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
	requestCacheTTL time.Duration,
	websocketService *websocket.Service,
	indexDBPath string,
	accountDumpsPath string,
	pub *publisher.Publisher,
	l1ParamsFetcher parameters.L1ParamsFetcher,
	jsonrpcParams *jsonrpc.Parameters,
) {
	// load mock files to generate correct echo swagger documentation
	mocker := NewMocker()
	mocker.LoadMockFiles()

	chainService := services.NewChainService(logger, chainsProvider, chainMetricsProvider, chainRecordRegistryProvider)
	committeeService := services.NewCommitteeService(chainsProvider, networkProvider, dkShareRegistryProvider)
	registryService := services.NewRegistryService(chainsProvider, chainRecordRegistryProvider)
	offLedgerService := services.NewOffLedgerService(chainService, networkProvider, requestCacheTTL)
	metricsService := services.NewMetricsService(chainsProvider, chainMetricsProvider)
	peeringService := services.NewPeeringService(chainsProvider, networkProvider, trustedNetworkManager)
	evmService := services.NewEVMService(chainsProvider, chainService, networkProvider, pub, indexDBPath, chainMetricsProvider, jsonrpcParams, logger.NewChildLogger("EVMService"))
	nodeService := services.NewNodeService(chainRecordRegistryProvider, nodeIdentityProvider, chainsProvider, shutdownHandler, trustedNetworkManager, l1ParamsFetcher)
	dkgService := services.NewDKGService(dkShareRegistryProvider, dkgNodeProvider, trustedNetworkManager)
	userService := services.NewUserService(userManager)
	// --

	authMiddleware := authentication.AddAuthentication(server, userManager, nodeIdentityProvider, authConfig, mocker)

	controllersToLoad := []interfaces.APIController{
		chain.NewChainController(logger, chainService, committeeService, evmService, nodeService, offLedgerService, registryService, accountDumpsPath),
		apimetrics.NewMetricsController(chainService, metricsService),
		node.NewNodeController(waspVersion, config, dkgService, nodeService, peeringService),
		requests.NewRequestsController(chainService, offLedgerService, peeringService),
		users.NewUsersController(userService),
		corecontracts.NewCoreContractsController(chainService),
	}

	AddHealthEndpoint(server, chainService, metricsService)
	addWebSocketEndpoint(server, websocketService)
	loadControllers(server, mocker, controllersToLoad, authMiddleware)
}
