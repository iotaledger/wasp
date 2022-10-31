package chain

import (
	"encoding/hex"
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
)

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
	chainNodeInfo, err := c.committeeService.GetCommitteeInfo(chain)

	if err != nil {
		return err
	}

	chainInfo := models.CommitteeInfoResponse{
		ChainID:        chainID.String(),
		Active:         chainRecord.Active,
		StateAddress:   committeeInfo.Address.String(),
		CommitteeNodes: models.MapCommitteeNodes(chainNodeInfo.CommitteeNodes),
		AccessNodes:    models.MapCommitteeNodes(chainNodeInfo.AccessNodes),
		CandidateNodes: models.MapCommitteeNodes(chainNodeInfo.CandidateNodes),
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

	chainList := make([]models.ChainInfoResponse, 0)

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

func (c *Controller) getState(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	stateKey, err := hex.DecodeString(e.Param("stateKey"))
	if err != nil {
		return apierrors.InvalidPropertyError("stateKey", err)
	}

	state, err := c.chainService.GetState(chainID, stateKey)

	if err != nil {
		return apierrors.InternalServerError(err)
	}

	return e.JSON(http.StatusOK, state)
}
