package metrics

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (c *Controller) getChainMessageMetrics(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	metricsReport := c.metricsService.GetChainMessageMetrics(ch.ID())
	mappedMetrics := models.MapChainMessageMetrics(metricsReport)

	return e.JSON(http.StatusOK, mappedMetrics)
}

func (c *Controller) getChainWorkflowMetrics(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	metricsReport := c.metricsService.GetChainConsensusWorkflowMetrics(ch.ID())

	return e.JSON(http.StatusOK, metricsReport)
}

func (c *Controller) getChainPipeMetrics(e echo.Context) error {
	ch, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	metricsReport := c.metricsService.GetChainConsensusPipeMetrics(ch.ID())

	return e.JSON(http.StatusOK, metricsReport)
}
