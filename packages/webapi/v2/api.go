package v2

import (
	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/services"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func Init(logger *loggerpkg.Logger,
	server echoswagger.ApiRoot,
	chainsProvider chains.Provider,
	metrics *metrics.Metrics,
	networkProvider peering.NetworkProvider,
	registryProvider registry.Provider,
	wal *wal.WAL,
) {
	server.SetRequestContentType(echo.MIMEApplicationJSON)
	server.SetResponseContentType(echo.MIMEApplicationJSON)

	vmService := services.NewVMService(logger, chainsProvider)
	chainService := services.NewChainService(logger, chainsProvider, metrics, registryProvider, vmService, wal)
	nodeService := services.NewNodeService(logger, networkProvider, registryProvider)
	registryService := services.NewRegistryService(logger, chainsProvider, metrics, registryProvider, wal)

	controllersToLoad := []interfaces.APIController{
		controllers.NewChainController(logger, chainService, nodeService, registryService),
	}

	publicRouter := server.Group("public", "v2").
		SetDescription("Public endpoints")

	adminRouter := server.Group("admin", "v2").
		SetDescription("Admin endpoints")

	for _, controller := range controllersToLoad {
		controller.RegisterPublic(publicRouter)
		controller.RegisterAdmin(adminRouter)
	}
}
