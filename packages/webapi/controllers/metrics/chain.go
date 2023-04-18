package metrics

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (c *Controller) getNodeMessageMetrics(e echo.Context) error {
	metricsReport := c.metricsService.GetNodeMessageMetrics()
	mappedMetrics := models.MapNodeMessageMetrics(metricsReport)

	return e.JSON(http.StatusOK, mappedMetrics)
}

func (c *Controller) getChainMessageMetrics(e echo.Context) error {
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	metricsReport := c.metricsService.GetChainMessageMetrics(chainID)
	mappedMetrics := models.MapChainMessageMetrics(metricsReport)

	return e.JSON(http.StatusOK, mappedMetrics)
}

func (c *Controller) getChainWorkflowMetrics(e echo.Context) error {
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	metricsReport := c.metricsService.GetChainConsensusWorkflowMetrics(chainID)

	return e.JSON(http.StatusOK, metricsReport)
}

func (c *Controller) getChainPipeMetrics(e echo.Context) error {
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	metricsReport := c.metricsService.GetChainConsensusPipeMetrics(chainID)

	return e.JSON(http.StatusOK, metricsReport)
}
