package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
)

//go:embed templates/metrics_chain.tmpl
var tplMetricsChain string

func metricsChainBreadcrumb(e *echo.Echo, chainID isc.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("metricsChain"),
		Title: fmt.Sprintf("Metrics: %.8s", chainID.String()),
		Href:  e.Reverse("metricsChain", chainID.String()),
	}
}

func (d *Dashboard) initMetricsChain(e *echo.Echo, r renderer) {
	route := e.GET("/metrics/:chainid", d.handleMetricsChain)
	route.Name = "metricsChain"
	r[route.Path] = d.makeTemplate(e, tplMetricsChain)
}

func (d *Dashboard) handleMetricsChain(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	tab := metricsChainBreadcrumb(c.Echo(), chainID)
	return c.Render(http.StatusOK, c.Path(), &MetricsChainTemplateParams{
		BaseTemplateParams: d.BaseParams(c, tab),
		ChainID:            chainID.String(),
	})
}

type MetricsChainTemplateParams struct {
	BaseTemplateParams
	ChainID string
}
