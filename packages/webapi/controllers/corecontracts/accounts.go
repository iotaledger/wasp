package corecontracts

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) getTotalAssets(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return c.handleViewCallError(err)
	}

	assets, err := corecontracts.GetTotalAssets(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err)
	}

	assetsResponse := &models.AssetsResponse{
		BaseTokens: assets.BaseTokens().String(),
		Coins:      assets.Coins.JSON(),
		Objects:    assets.Objects.JSON(),
	}

	return e.JSON(http.StatusOK, assetsResponse)
}

func (c *Controller) getAccountBalance(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return c.handleViewCallError(err)
	}

	agentID, err := params.DecodeAgentID(e)
	if err != nil {
		return err
	}

	assets, err := corecontracts.GetAccountBalance(ch, agentID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err)
	}

	assetsResponse := &models.AssetsResponse{
		BaseTokens: assets.BaseTokens().String(),
		Coins:      assets.Coins.JSON(),
		Objects:    assets.Objects.JSON(),
	}

	return e.JSON(http.StatusOK, assetsResponse)
}

func (c *Controller) getAccountFoundries(e echo.Context) error {
	panic("TODO")
	// ch,  err := c.chainService.GetChain()
	// if err != nil {
	// 	return c.handleViewCallError(err)
	// }
	// agentID, err := params.DecodeAgentID(e)
	// if err != nil {
	// 	return err
	// }

	// foundries, err := corecontracts.GetAccountFoundries(ch, agentID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	// if err != nil {
	// 	return c.handleViewCallError(err)
	// }

	// return e.JSON(http.StatusOK, &models.AccountFoundriesResponse{
	// 	FoundrySerialNumbers: foundries,
	// })
}

func (c *Controller) getAccountNonce(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return c.handleViewCallError(err)
	}

	agentID, err := params.DecodeAgentID(e)
	if err != nil {
		return err
	}

	nonce, err := corecontracts.GetAccountNonce(ch, agentID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err)
	}

	nonceResponse := &models.AccountNonceResponse{
		Nonce: fmt.Sprint(nonce),
	}

	return e.JSON(http.StatusOK, nonceResponse)
}

// func (c *Controller) getNativeTokenIDRegistry(e echo.Context) error {
// 	ch, err := c.chainService.GetChain()
// 	if err != nil {
// 		return c.handleViewCallError(err)
// 	}

// 	registries, err := corecontracts.GetNativeTokenIDRegistry(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
// 	if err != nil {
// 		return c.handleViewCallError(err)
// 	}

// 	nativeTokenIDRegistryResponse := &models.NativeTokenIDRegistryResponse{
// 		NativeTokenRegistryIDs: make([]string, len(registries)),
// 	}

// 	for k, v := range registries {
// 		nativeTokenIDRegistryResponse.NativeTokenRegistryIDs[k] = v.String()
// 	}

// 	return e.JSON(http.StatusOK, nativeTokenIDRegistryResponse)
// }

func (c *Controller) getFoundryOutput(e echo.Context) error {
	panic("TODO")
	// ch,  err := c.chainService.GetChain()
	// if err != nil {
	// 	return c.handleViewCallError(err)
	// }

	// serialNumber, err := params.DecodeUInt(e, "serialNumber")
	// if err != nil {
	// 	return err
	// }

	// foundryOutput, err := corecontracts.GetFoundryOutput(ch, uint32(serialNumber), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	// if err != nil {
	// 	return c.handleViewCallError(err)
	// }

	// foundryOutputID, err := foundryOutput.ID()
	// if err != nil {
	// 	return apierrors.InvalidPropertyError("FoundryOutput.ID", err)
	// }

	// foundryOutputResponse := &models.FoundryOutputResponse{
	// 	FoundryID: foundryOutputID.ToHex(),
	// 	Assets: models.AssetsResponse{
	// 		BaseTokens:   iotago.EncodeUint64(foundryOutput.Amount),
	// 		NativeTokens: isc.NativeTokensToJSONObject(foundryOutput.NativeTokens),
	// 	},
	// }

	// return e.JSON(http.StatusOK, foundryOutputResponse)
}
