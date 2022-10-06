package metrics

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
)

func (c *Controller) getChainMetrics(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	metricsReport := c.metricsService.GetConnectionMetrics(chainID)

	return e.JSON(http.StatusOK, metricsReport)
}
