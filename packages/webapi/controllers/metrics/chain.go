package metrics

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

func (c *Controller) getChainMessageMetrics(e echo.Context) error {
	metricsReport := c.metricsService.GetChainMessageMetrics()
	mappedMetrics := models.MapChainMessageMetrics(metricsReport)

	return e.JSON(http.StatusOK, mappedMetrics)
}

func (c *Controller) getChainWorkflowMetrics(e echo.Context) error {
	metricsReport := c.metricsService.GetChainConsensusWorkflowMetrics()

	return e.JSON(http.StatusOK, metricsReport)
}

func (c *Controller) getChainPipeMetrics(e echo.Context) error {
	metricsReport := c.metricsService.GetChainConsensusPipeMetrics()

	return e.JSON(http.StatusOK, metricsReport)
}
