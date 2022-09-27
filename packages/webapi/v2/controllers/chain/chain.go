package chain

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
)

func (c *Controller) activateChain(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	if err := c.chainService.ActivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) deactivateChain(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	if err := c.chainService.DeactivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) getCommitteeInfo(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	chainRecord, err := c.registryService.GetChainRecordByChainID(chainID)
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	if chainRecord == nil {
		return apierrors.ChainNotFoundError(e.Param("chainID"))
	}

	chain := c.chainService.GetChainByID(chainID)

	if chain == nil {
		return apierrors.ChainNotFoundError(e.Param("chainID"))
	}

	committeeInfo := chain.GetCommitteeInfo()
	chainNodeInfo, err := c.nodeService.GetNodeInfo(chain)
	if err != nil {
		return err
	}

	chainInfo := models.CommitteeInfoResponse{
		ChainID:        chainID.String(),
		Active:         chainRecord.Active,
		StateAddress:   committeeInfo.Address.String(),
		CommitteeNodes: chainNodeInfo.CommitteeNodes,
		AccessNodes:    chainNodeInfo.AccessNodes,
		CandidateNodes: chainNodeInfo.CandidateNodes,
	}

	return e.JSON(http.StatusOK, chainInfo)
}

func (c *Controller) getChainInfo(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	chainInfo, err := c.chainService.GetChainInfoByChainID(chainID)
	if err != nil {
		return err
	}

	evmChainID, err := c.chainService.GetEVMChainID(chainID)
	if err != nil {
		return err
	}

	chainInfoResponse := models.MapChainInfoResponse(chainInfo, evmChainID)

	return e.JSON(http.StatusOK, chainInfoResponse)
}

func (c *Controller) getChainList(e echo.Context) error {
	chainIDs, err := c.chainService.GetAllChainIDs()
	if err != nil {
		return err
	}

	chainList := models.ChainListResponse{}

	for _, chainID := range chainIDs {
		chainInfo, err := c.chainService.GetChainInfoByChainID(chainID)
		if err != nil {
			return err
		}

		evmChainID, err := c.chainService.GetEVMChainID(chainID)
		if err != nil {
			return err
		}

		chainInfoResponse := models.MapChainInfoResponse(chainInfo, evmChainID)

		chainList = append(chainList, chainInfoResponse)
	}

	return e.JSON(http.StatusOK, chainList)
}
