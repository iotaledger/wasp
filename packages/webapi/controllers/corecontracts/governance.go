package corecontracts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func MapGovChainInfoResponse(chainInfo *isc.ChainInfo) models.GovChainInfoResponse {
	return models.GovChainInfoResponse{
		ChainID:      chainInfo.ChainID.String(),
		ChainOwnerID: chainInfo.ChainOwnerID.String(),
		GasFeePolicy: chainInfo.GasFeePolicy,
		GasLimits:    chainInfo.GasLimits,
		PublicURL:    chainInfo.PublicURL,
		Metadata: models.GovPublicChainMetadata{
			EVMJsonRPCURL:   chainInfo.Metadata.EVMJsonRPCURL,
			EVMWebSocketURL: chainInfo.Metadata.EVMWebSocketURL,
			Name:            chainInfo.Metadata.Name,
			Description:     chainInfo.Metadata.Description,
			Website:         chainInfo.Metadata.Website,
		},
	}
}

func (c *Controller) getChainInfo(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	chainInfo, err := corecontracts.GetChainInfo(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	chainInfoResponse := MapGovChainInfoResponse(chainInfo)

	return e.JSON(http.StatusOK, chainInfoResponse)
}

func (c *Controller) getChainOwner(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	chainOwner, err := corecontracts.GetChainOwner(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	chainOwnerResponse := models.GovChainOwnerResponse{
		ChainOwner: chainOwner.String(),
	}

	return e.JSON(http.StatusOK, chainOwnerResponse)
}

func (c *Controller) getAllowedStateControllerAddresses(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	addresses, err := corecontracts.GetAllowedStateControllerAddresses(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
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
