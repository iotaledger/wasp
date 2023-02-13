package chain

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) getCommitteeInfo(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	chain, err := c.chainService.GetChainInfoByChainID(chainID)
	if err != nil {
		return apierrors.ChainNotFoundError(chainID.String())
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

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	chainInfo, err := c.chainService.GetChainInfoByChainID(chainID)
	if errors.Is(err, interfaces.ErrChainNotFound) {
		return e.NoContent(http.StatusNotFound)
	} else if err != nil {
		return err
	}

	evmChainID := uint16(0)
	if chainInfo.IsActive {
		evmChainID, err = c.chainService.GetEVMChainID(chainID)
		if err != nil {
			return err
		}
	}

	chainInfoResponse := models.MapChainInfoResponse(chainInfo, evmChainID)

	return e.JSON(http.StatusOK, chainInfoResponse)
}

func (c *Controller) getChainList(e echo.Context) error {
	chainIDs, err := c.chainService.GetAllChainIDs()
	c.log.Info("After allChainIDS %v", err)
	if err != nil {
		return err
	}

	chainList := make([]models.ChainInfoResponse, 0)

	for _, chainID := range chainIDs {
		chainInfo, err := c.chainService.GetChainInfoByChainID(chainID)
		c.log.Info("getchaininfo %v", err)

		if errors.Is(err, interfaces.ErrChainNotFound) {
			// TODO: Validate this logic here. Is it possible to still get more chain info?
			chainList = append(chainList, models.ChainInfoResponse{
				IsActive: false,
				ChainID:  chainID.String(),
			})
			continue
		} else if err != nil {
			return err
		}

		evmChainID := uint16(0)
		if chainInfo.IsActive {
			evmChainID, err = c.chainService.GetEVMChainID(chainID)
			c.log.Info("getevmchainid %v", err)

			if err != nil {
				return err
			}
		}

		chainInfoResponse := models.MapChainInfoResponse(chainInfo, evmChainID)
		c.log.Info("mapchaininfo %v", err)

		chainList = append(chainList, chainInfoResponse)
	}

	return e.JSON(http.StatusOK, chainList)
}

func (c *Controller) getState(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	stateKey, err := iotago.DecodeHex(e.Param("stateKey"))
	if err != nil {
		return apierrors.InvalidPropertyError("stateKey", err)
	}

	state, err := c.chainService.GetState(chainID, stateKey)
	if err != nil {
		return apierrors.InternalServerError(err)
	}

	response := models.StateResponse{
		State: iotago.EncodeHex(state),
	}

	return e.JSON(http.StatusOK, response)
}
