package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/labstack/echo/v4"
)

//go:embed templates/metrics_chain.tmpl
var tplMetricsChain string

func metricsChainBreadcrumb(e *echo.Echo, chainID *iscp.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("metricsChain"),
		Title: fmt.Sprintf("Metrics: %.8s", chainID.Base58()),
		Href:  e.Reverse("metricsChain", chainID.Base58()),
	}
}

func (d *Dashboard) initMetricsChain(e *echo.Echo, r renderer) {
	route := e.GET("/metrics/:chainid", d.handleMetricsChain)
	route.Name = "metricsChain"
	r[route.Path] = d.makeTemplate(e, tplMetricsChain)
}

func (d *Dashboard) handleMetricsChain(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	tab := metricsChainBreadcrumb(c.Echo(), chainID)
	return c.Render(http.StatusOK, c.Path(), &MetricsChainTemplateParams{
		BaseTemplateParams: d.BaseParams(c, tab),
		ChainID:            chainID.Base58(),
	})
}

type MetricsChainTemplateParams struct {
	BaseTemplateParams
	ChainID string
}
