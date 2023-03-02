package corecontracts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func MapGovChainInfoResponse(chainInfo *governance.ChainInfo) models.GovChainInfoResponse {
	return models.GovChainInfoResponse{
		ChainID:      chainInfo.ChainID.String(),
		ChainOwnerID: chainInfo.ChainOwnerID.String(),
		Description:  chainInfo.Description,
		GasFeePolicy: models.GasFeePolicy{
			GasPerToken:       chainInfo.GasFeePolicy.GasPerToken,
			ValidatorFeeShare: chainInfo.GasFeePolicy.ValidatorFeeShare,
			EVMGasRatio:       chainInfo.GasFeePolicy.EVMGasRatio,
		},
		MaxBlobSize:     chainInfo.MaxBlobSize,
		MaxEventSize:    chainInfo.MaxEventSize,
		MaxEventsPerReq: chainInfo.MaxEventsPerReq,
	}
}

func (c *Controller) getChainInfo(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	chainInfo, err := c.governance.GetChainInfo(chainID)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	chainInfoResponse := MapGovChainInfoResponse(chainInfo)

	return e.JSON(http.StatusOK, chainInfoResponse)
}

func (c *Controller) getChainOwner(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	chainOwner, err := c.governance.GetChainOwner(chainID)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	chainOwnerResponse := models.GovChainOwnerResponse{
		ChainOwner: chainOwner.String(),
	}

	return e.JSON(http.StatusOK, chainOwnerResponse)
}

func (c *Controller) getAllowedStateControllerAddresses(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	addresses, err := c.governance.GetAllowedStateControllerAddresses(chainID)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	encodedAddresses := make([]string, len(addresses))

	for k, v := range addresses {
		encodedAddresses[k] = v.Bech32(parameters.L1().Protocol.Bech32HRP)
	}

	addressesResponse := models.GovAllowedStateControllerAddressesResponse{
		Addresses: encodedAddresses,
	}

	return e.JSON(http.StatusOK, addressesResponse)
}
