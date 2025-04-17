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
		Coins: assets.JSON(),
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
		Coins: assets.JSON(),
	}

	return e.JSON(http.StatusOK, assetsResponse)
}

func (c *Controller) getAccountObjects(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return c.handleViewCallError(err)
	}

	agentID, err := params.DecodeAgentID(e)
	if err != nil {
		return err
	}

	objects, err := corecontracts.GetAccountObjects(ch, agentID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return c.handleViewCallError(err)
	}

	objectsResponse := &models.AccountObjectsResponse{
		ObjectIDs:   make([]string, len(objects)),
		ObjectTypes: make([]string, len(objects)),
	}

	for k, v := range objects {
		objectsResponse.ObjectIDs[k] = v.ID.String()
		objectsResponse.ObjectTypes[k] = v.Type.String()
	}

	return e.JSON(http.StatusOK, objectsResponse)
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
