package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/labstack/echo/v4"
)

//go:embed templates/metrics_chain_nodeconn.tmpl
var tplMetricsChainNodeconn string

func metricsChainNodeconnBreadcrumb(e *echo.Echo, chainID *isc.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("metricsChainNodeconn"),
		Title: fmt.Sprintf("Metrics: %.8s: Connection to L1", chainID.String()),
		Href:  e.Reverse("metricsChainNodeconn", chainID.String()),
	}
}

func (d *Dashboard) initMetricsChainNodeconn(e *echo.Echo, r renderer) {
	route := e.GET("/metrics/:chainid/nodeconn", d.handleMetricsChainNodeconn)
	route.Name = "metricsChainNodeconn"
	r[route.Path] = d.makeTemplate(e, tplMetricsChainNodeconn)
}

func (d *Dashboard) handleMetricsChainNodeconn(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	tab := metricsChainNodeconnBreadcrumb(c.Echo(), chainID)
	metrics, err := d.wasp.GetChainNodeConnectionMetrics(chainID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	return c.Render(http.StatusOK, c.Path(), &MetricsChainNodeconnTemplateParams{
		BaseTemplateParams: d.BaseParams(c, metricsChainBreadcrumb(c.Echo(), chainID), tab),
		ChainID:            chainID.String(),
		Metrics:            metrics,
	})
}

type MetricsChainNodeconnTemplateParams struct {
	BaseTemplateParams
	ChainID string
	Metrics nodeconnmetrics.NodeConnectionMessagesMetrics
}
