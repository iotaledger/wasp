package metrics

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (c *Controller) getHealth(e echo.Context) error {
	metricsReport := c.metricsService.GetNodeMessageMetrics()
	mappedMetrics := models.MapNodeMessageMetrics(metricsReport).ChainConfirmedState

	return e.JSON(http.StatusOK, mappedMetrics)
}
