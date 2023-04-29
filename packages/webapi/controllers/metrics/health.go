package metrics

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (c *Controller) getHealth(e echo.Context) error {
	metricsReport := c.metricsService.GetNodeMessageMetrics()
	mappedMetrics := models.MapNodeMessageMetrics(metricsReport).ChainConfirmedState

	if mappedMetrics.LastMessage.ConfirmedStateWant-mappedMetrics.LastMessage.ConfirmedStateHave < 2 {
		return e.JSON(http.StatusOK, mappedMetrics)
	}
	return e.JSON(http.StatusInternalServerError, mappedMetrics)
}
