package metrics

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (c *Controller) getHealth(e echo.Context) error {
	chainIDs, err := c.chainService.GetAllChainIDs()
	if err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("failed to get all chain IDs: %s", err))
	}

	for _, chainID := range chainIDs {
		lag := c.metricsService.GetMaxChainConfirmedStateLag(chainID)
		if lag > 2 {
			return e.String(http.StatusInternalServerError, fmt.Sprintf("chain %v not sync with %d diff", chainID.String(), lag))
		}
	}
	return e.String(http.StatusOK, "all chain synchronized")
}
