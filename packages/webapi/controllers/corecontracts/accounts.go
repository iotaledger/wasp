package corecontracts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) getAccounts(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	accounts, err := corecontracts.GetAccounts(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	accountsResponse := &models.AccountListResponse{
		Accounts: make([]string, len(accounts)),
	}

	for k, v := range accounts {
		accountsResponse.Accounts[k] = v.String()
	}

	return e.JSON(http.StatusOK, accountsResponse)
}

func (c *Controller) getTotalAssets(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	assets, err := corecontracts.GetTotalAssets(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	assetsResponse := &models.AssetsResponse{
		BaseTokens:   iotago.EncodeUint64(assets.BaseTokens),
		NativeTokens: models.MapNativeTokens(assets.NativeTokens),
	}

	return e.JSON(http.StatusOK, assetsResponse)
}

func (c *Controller) getAccountBalance(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	agentID, err := params.DecodeAgentID(e)
	if err != nil {
		return err
	}

	assets, err := corecontracts.GetAccountBalance(ch, agentID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	assetsResponse := &models.AssetsResponse{
		BaseTokens:   iotago.EncodeUint64(assets.BaseTokens),
		NativeTokens: models.MapNativeTokens(assets.NativeTokens),
	}

	return e.JSON(http.StatusOK, assetsResponse)
}

func (c *Controller) getAccountNFTs(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	agentID, err := params.DecodeAgentID(e)
	if err != nil {
		return err
	}

	nfts, err := corecontracts.GetAccountNFTs(ch, agentID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	nftsResponse := &models.AccountNFTsResponse{
		NFTIDs: make([]string, len(nfts)),
	}

	for k, v := range nfts {
		nftsResponse.NFTIDs[k] = v.ToHex()
	}

	return e.JSON(http.StatusOK, nftsResponse)
}

func (c *Controller) getAccountFoundries(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}
	agentID, err := params.DecodeAgentID(e)
	if err != nil {
		return err
	}

	foundries, err := corecontracts.GetAccountFoundries(ch, agentID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	return e.JSON(http.StatusOK, &models.AccountFoundriesResponse{
		FoundrySerialNumbers: foundries,
	})
}

func (c *Controller) getAccountNonce(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	agentID, err := params.DecodeAgentID(e)
	if err != nil {
		return err
	}

	nonce, err := corecontracts.GetAccountNonce(ch, agentID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	nonceResponse := &models.AccountNonceResponse{
		Nonce: iotago.EncodeUint64(nonce),
	}

	return e.JSON(http.StatusOK, nonceResponse)
}

func (c *Controller) getNFTData(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	nftID, err := params.DecodeNFTID(e)
	if err != nil {
		return err
	}

	nftData, err := corecontracts.GetNFTData(ch, *nftID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	nftDataResponse := models.MapNFTDataResponse(nftData)

	return e.JSON(http.StatusOK, nftDataResponse)
}

func (c *Controller) getNativeTokenIDRegistry(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	registries, err := corecontracts.GetNativeTokenIDRegistry(ch, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	nativeTokenIDRegistryResponse := &models.NativeTokenIDRegistryResponse{
		NativeTokenRegistryIDs: make([]string, len(registries)),
	}

	for k, v := range registries {
		nativeTokenIDRegistryResponse.NativeTokenRegistryIDs[k] = v.String()
	}

	return e.JSON(http.StatusOK, nativeTokenIDRegistryResponse)
}

func (c *Controller) getFoundryOutput(e echo.Context) error {
	ch, chainID, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	serialNumber, err := params.DecodeUInt(e, "serialNumber")
	if err != nil {
		return err
	}

	foundryOutput, err := corecontracts.GetFoundryOutput(ch, uint32(serialNumber), e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err, chainID)
	}

	foundryOutputID, err := foundryOutput.ID()
	if err != nil {
		return apierrors.InvalidPropertyError("FoundryOutput.ID", err)
	}

	foundryOutputResponse := &models.FoundryOutputResponse{
		FoundryID: foundryOutputID.ToHex(),
		Assets: models.AssetsResponse{
			BaseTokens:   iotago.EncodeUint64(foundryOutput.Amount),
			NativeTokens: models.MapNativeTokens(foundryOutput.NativeTokens),
		},
	}

	return e.JSON(http.StatusOK, foundryOutputResponse)
}
