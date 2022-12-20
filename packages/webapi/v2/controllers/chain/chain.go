package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
	"github.com/iotaledger/wasp/packages/webapi/v2/params"
)

func (c *Controller) getCommitteeInfo(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	chain, err := c.chainService.GetChainInfoByChainID(chainID)
	if err != nil {
		return apierrors.ChainNotFoundError(e.Param("chainID"))
	}

	chainNodeInfo, err := c.committeeService.GetCommitteeInfo(chainID)
	if err != nil {
		return err
	}

	chainInfo := models.CommitteeInfoResponse{
		ChainID:        chainID.String(),
		Active:         chain.IsActive,
		StateAddress:   chainNodeInfo.Address.String(),
		CommitteeNodes: models.MapCommitteeNodes(chainNodeInfo.CommitteeNodes),
		AccessNodes:    models.MapCommitteeNodes(chainNodeInfo.AccessNodes),
		CandidateNodes: models.MapCommitteeNodes(chainNodeInfo.CandidateNodes),
	}

	return e.JSON(http.StatusOK, chainInfo)
}

func (c *Controller) getChainInfo(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
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
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	stateKey, err := iotago.DecodeHex(e.Param("stateKey"))
	if err != nil {
		return apierrors.InvalidPropertyError("stateKey", err)
	}

	state, err := c.chainService.GetState(chainID, stateKey)
	if err != nil {
		return apierrors.InternalServerError(err)
	}

	return e.JSON(http.StatusOK, state)
}
