// Package corecontracts contains routes for webapi core contract interactions
package corecontracts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/packages/webapi/params"
)

func MapGovChainInfoResponse(chainInfo *isc.ChainInfo) models.GovChainInfoResponse {
	return models.GovChainInfoResponse{
		ChainID:      chainInfo.ChainID.String(),
		ChainAdmin:   chainInfo.ChainAdmin.String(),
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
	ch, err := c.chainService.GetChain()
	if err != nil {
		return c.handleViewCallError(err)
	}

	chainInfo, err := corecontracts.GetChainInfo(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err)
	}

	chainInfoResponse := MapGovChainInfoResponse(chainInfo)

	return e.JSON(http.StatusOK, chainInfoResponse)
}

func (c *Controller) getChainAdmin(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return c.handleViewCallError(err)
	}

	chainAdmin, err := corecontracts.GetChainAdmin(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err)
	}

	chainAdminResponse := models.GovChainAdminResponse{
		ChainAdmin: chainAdmin.String(),
	}
	return e.JSON(http.StatusOK, chainAdminResponse)
}
