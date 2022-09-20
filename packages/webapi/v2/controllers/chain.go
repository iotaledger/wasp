package controllers

import (
	"net/http"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/models"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type ChainController struct {
	logger          *logger.Logger
	chainService    interfaces.Chain
	nodeService     interfaces.Node
	registryService interfaces.Registry
}

func NewChainController(logger *logger.Logger, chainService interfaces.Chain, nodeService interfaces.Node, registryService interfaces.Registry) interfaces.APIController {
	return &ChainController{
		logger:          logger,
		chainService:    chainService,
		nodeService:     nodeService,
		registryService: registryService,
	}
}

func (c *ChainController) activateChain(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return err
	}

	if err := c.chainService.ActivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *ChainController) deactivateChain(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return err
	}

	if err := c.chainService.DeactivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *ChainController) getChainInfo(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return err
	}

	chainRecord, err := c.registryService.GetChainRecordByChainID(chainID)
	if err != nil {
		return err
	}

	if chainRecord == nil {
		return httperrors.NotFound("")
	}

	chain := c.chainService.GetChainByID(chainID)

	if chain == nil {
		return httperrors.NotFound("")
	}

	committeeInfo := chain.GetCommitteeInfo()
	chainNodeInfo, err := c.nodeService.GetNodeInfo(chain)
	if err != nil {
		return err
	}

	chainInfo := models.ChainInfo{
		ChainID:        chainID.String(),
		Active:         chainRecord.Active,
		StateAddress:   committeeInfo.Address.String(),
		CommitteeNodes: chainNodeInfo.CommitteeNodes,
		AccessNodes:    chainNodeInfo.AccessNodes,
		CandidateNodes: chainNodeInfo.CandidateNodes,
	}

	return e.JSON(http.StatusOK, chainInfo)
}

func (c *ChainController) getContracts(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return err
	}

	contracts, err := c.chainService.GetContracts(chainID)
	if err != nil {
		return err
	}

	contractList := make([]models.ContractInfo, 0, len(contracts))

	for hname, contract := range contracts {
		contractInfo := models.ContractInfo{
			Description: contract.Description,
			HName:       hname,
			Name:        contract.Name,
			ProgramHash: contract.ProgramHash,
		}

		contractList = append(contractList, contractInfo)
	}

	return e.JSON(http.StatusOK, contractList)
}

func (c *ChainController) RegisterPublic(publicAPI echoswagger.ApiGroup) {
}

func (c *ChainController) RegisterAdmin(adminAPI echoswagger.ApiGroup) {
	adminAPI.POST(routes.ActivateChain(":chainID"), c.activateChain).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Activate a chain").
		SetOperationId("activateChain")

	adminAPI.POST(routes.DeactivateChain(":chainID"), c.deactivateChain).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Deactivate a chain")

	adminAPI.GET(routes.GetChainInfo(":chainID"), c.getChainInfo).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Get basic chain info.")

	adminAPI.GET(routes.GetChainContracts(":chainID"), c.getContracts).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Get all available chain contracts.")
}
