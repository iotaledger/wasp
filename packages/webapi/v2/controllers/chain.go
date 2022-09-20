package controllers

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/models"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"net/http"
)

type ChainController struct {
	logger          *logger.Logger
	chainService    interfaces.Chain
	nodeService interfaces.Node
	registryService interfaces.Registry
}

func NewChainController(logger *logger.Logger, chainService interfaces.Chain, nodeService interfaces.Node, registryService interfaces.Registry) interfaces.APIController {
	return &ChainController{
		logger:       logger,
		chainService: chainService,
		nodeService: nodeService,
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
		ChainID:        chainID,
		Active:         chainRecord.Active,
		StateAddress:   committeeInfo.Address,
		CommitteeNodes: chainNodeInfo.CommitteeNodes,
		AccessNodes:    chainNodeInfo.AccessNodes,
		CandidateNodes: chainNodeInfo.CandidateNodes,
	}

	return e.JSON(http.StatusOK, chainInfo)
}

func (c *ChainController) RegisterPublic(server echoswagger.ApiRoot) {

}

func (c *ChainController) RegisterAdmin(server echoswagger.ApiRoot) {
	server.POST(routes.ActivateChain(":chainID"), c.activateChain).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Activate a chain")

	server.POST(routes.DeactivateChain(":chainID"), c.deactivateChain).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Deactivate a chain")

	server.GET(routes.GetChainInfo(":chainID"), c.getChainInfo).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Get basic chain info.")
}
