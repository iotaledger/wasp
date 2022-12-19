package v2

import (
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/hive.go/core/configuration"
	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	userspkg "github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/chain"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/metrics"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/node"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/requests"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/users"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/services"
)

func loadControllers(server echoswagger.ApiRoot, userManager *userspkg.UserManager, nodeIdentityProvider registry.NodeIdentityProvider, authConfig authentication.AuthConfiguration, mocker *Mocker, controllersToLoad []interfaces.APIController) {
	for _, controller := range controllersToLoad {
		publicGroup := server.Group(controller.Name(), "v2/")

		controller.RegisterPublic(publicGroup, mocker)

		claimValidator := func(claims *authentication.WaspClaims) bool {
			// The API will be accessible if the token has an 'API' claim
			return claims.HasPermission(permissions.API)
		}

		adminGroup := server.Group(controller.Name(), "v2/").
			SetSecurity("Authorization")

		authentication.AddAuthentication(adminGroup.EchoGroup(), userManager, nodeIdentityProvider, authConfig, claimValidator)

		controller.RegisterAdmin(adminGroup, mocker)
	}
}

func Init(
	logger *loggerpkg.Logger,
	server echoswagger.ApiRoot,
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
	nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics,
	authConfig authentication.AuthConfiguration,
	nodeOwnerAddresses []string,
) {
	mocker := NewMocker()
	mocker.LoadMockFiles()

	// -- Add dependency injection here
	vmService := services.NewVMService(chainsProvider)
	chainService := services.NewChainService(chainsProvider, nodeConnectionMetrics, chainRecordRegistryProvider, vmService)
	committeeService := services.NewCommitteeService(chainsProvider, networkProvider, dkShareRegistryProvider)
	registryService := services.NewRegistryService(chainsProvider, chainRecordRegistryProvider)
	offLedgerService := services.NewOffLedgerService(chainService, networkProvider)
	metricsService := services.NewMetricsService(chainsProvider)
	peeringService := services.NewPeeringService(chainsProvider, networkProvider, trustedNetworkManager)
	evmService := services.NewEVMService(chainService, networkProvider)
	nodeService := services.NewNodeService(nodeOwnerAddresses, nodeIdentityProvider, shutdownHandler)
	dkgService := services.NewDKGService(dkShareRegistryProvider, dkgNodeProvider)
	userService := services.NewUserService(userManager)
	// --

	controllersToLoad := []interfaces.APIController{
		chain.NewChainController(logger, chainService, committeeService, evmService, offLedgerService, registryService, vmService),
		metrics.NewMetricsController(metricsService),
		node.NewNodeController(config, dkgService, nodeService, peeringService),
		requests.NewRequestsController(chainService, offLedgerService, peeringService, vmService),
		users.NewUsersController(userService),
		corecontracts.NewCoreContractsController(vmService),
	}

	loadControllers(server, userManager, nodeIdentityProvider, authConfig, mocker, controllersToLoad)
}
