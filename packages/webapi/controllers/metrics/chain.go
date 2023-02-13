package metrics

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) getL1Metrics(e echo.Context) error {
	metricsReport := c.metricsService.GetAllChainsMetrics()
	mappedMetrics := models.MapChainMetrics(metricsReport)

	return e.JSON(http.StatusOK, mappedMetrics)
}

func (c *Controller) getChainMetrics(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	metricsReport := c.metricsService.GetChainMetrics(chainID)
	mappedMetrics := models.MapChainMetrics(metricsReport)

	return e.JSON(http.StatusOK, mappedMetrics)
}

func (c *Controller) getChainWorkflowMetrics(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	metricsReport := c.metricsService.GetChainConsensusWorkflowMetrics(chainID)

	return e.JSON(http.StatusOK, metricsReport)
}

func (c *Controller) getChainPipeMetrics(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	metricsReport := c.metricsService.GetChainConsensusPipeMetrics(chainID)

	return e.JSON(http.StatusOK, metricsReport)
}
