package chain

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/routes"
)

type Controller struct {
	log *loggerpkg.Logger

	chainService     interfaces.Chain
	nodeService      interfaces.Node
	offLedgerService interfaces.OffLedger
	registryService  interfaces.Registry
	vmService        interfaces.VM
}

func NewMetricsController(log *loggerpkg.Logger, chainService interfaces.Chain, nodeService interfaces.Node, offLedgerService interfaces.OffLedger, registryService interfaces.Registry, vmService interfaces.VM) interfaces.APIController {
	return &Controller{
		log:              log,
		chainService:     chainService,
		nodeService:      nodeService,
		offLedgerService: offLedgerService,
		registryService:  registryService,
		vmService:        vmService,
	}
}

func (c *Controller) Name() string {
	return "metrics"
}

func (c *Controller) RegisterExampleData(mock interfaces.Mocker) {
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET(routes.GetChainContracts(":chainID"), c.getChainMetrics).
		AddParamPath("", "chainID", "ChainID (Bech32)").
		AddResponse(http.StatusOK, "A list of all available contracts.", mocker.GetMockedStruct(models.ContractListResponse{}), nil).
		SetOperationId("getChainContracts").
		SetResponseContentType("application/json").
		SetSummary("Get all available chain contracts.")
}
