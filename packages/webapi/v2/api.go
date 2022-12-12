package v2

import (
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	corecontracts "github.com/iotaledger/wasp/packages/webapi/v2/controllers/core_contracts"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/metrics"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/wasp/packages/dkg"
	userspkg "github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/requests"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/users"

	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/node"

	metricspkg "github.com/iotaledger/wasp/packages/metrics"

	"github.com/iotaledger/hive.go/core/configuration"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/chain"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	walpkg "github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/services"
)

func loadControllers(server echoswagger.ApiRoot, mocker *Mocker, controllersToLoad []interfaces.APIController) {
	for _, controller := range controllersToLoad {
		publicGroup := server.Group(controller.Name(), "v2/")

		controller.RegisterPublic(publicGroup, mocker)

		adminGroup := server.Group(controller.Name(), "v2/").
			SetSecurity("Authorization")

		controller.RegisterAdmin(adminGroup, mocker)
	}
}

func Init(logger *loggerpkg.Logger,
	server echoswagger.ApiRoot,
	authConfig authentication.AuthConfiguration,
	config *configuration.Configuration,
	chainsProvider chains.Provider,
	dkgNodeProvider dkg.NodeProvider,
	metricsProvider *metricspkg.Metrics,
	networkProvider peering.NetworkProvider,
	nodeOwnerAddresses []string,
	registryProvider registry.Provider,
	shutdownHandler *shutdown.ShutdownHandler,
	trustedNetworkManager peering.TrustedNetworkManager,
	userManager *userspkg.UserManager,
	wal *walpkg.WAL,
) {
	mocker := NewMocker()
	mocker.LoadMockFiles()

	// -- Add dependency injection here
	vmService := services.NewVMService(logger, chainsProvider)
	chainService := services.NewChainService(logger, chainsProvider, metricsProvider, registryProvider, vmService, wal)
	committeeService := services.NewCommitteeService(logger, chainsProvider, networkProvider, registryProvider)
	registryService := services.NewRegistryService(logger, chainsProvider, registryProvider)
	offLedgerService := services.NewOffLedgerService(logger, chainService, networkProvider)
	metricsService := services.NewMetricsService(logger, chainsProvider)
	peeringService := services.NewPeeringService(logger, chainsProvider, networkProvider, trustedNetworkManager)
	evmService := services.NewEVMService(logger, chainService, networkProvider)
	nodeService := services.NewNodeService(logger, nodeOwnerAddresses, registryProvider, shutdownHandler)
	dkgService := services.NewDKGService(logger, registryProvider, dkgNodeProvider)
	userService := services.NewUserService(logger, userManager)
	// --

	controllersToLoad := []interfaces.APIController{
		chain.NewChainController(logger, chainService, committeeService, evmService, offLedgerService, registryService, vmService),
		metrics.NewMetricsController(logger, metricsService),
		node.NewNodeController(logger, config, dkgService, nodeService, peeringService),
		requests.NewRequestsController(logger, chainService, offLedgerService, peeringService, vmService),
		users.NewUsersController(logger, userService),
		corecontracts.NewCoreContractsController(logger, vmService),
	}

	claimValidator := func(claims *authentication.WaspClaims) bool {
		// The API will be accessible if the token has an 'API' claim
		return claims.HasPermission(permissions.API)
	}

	authentication.AddAuthentication(server.Echo(), userManager, registryProvider, authConfig, claimValidator)

	loadControllers(server, mocker, controllersToLoad)
}
