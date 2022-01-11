package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/labstack/echo/v4"
)

//go:embed templates/metrics_chain_nodeconn.tmpl
var tplMetricsChainNodeconn string

func metricsChainNodeconnBreadcrumb(e *echo.Echo, chainID *iscp.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("metricsChainNodeconn"),
		Title: fmt.Sprintf("Metrics: %.8s: Connection to L1", chainID.Base58()),
		Href:  e.Reverse("metricsChainNodeconn", chainID.Base58()),
	}
}

func (d *Dashboard) initMetricsChainNodeconn(e *echo.Echo, r renderer) {
	route := e.GET("/metrics/:chainid/nodeconn", d.handleMetricsChainNodeconn)
	route.Name = "metricsChainNodeconn"
	r[route.Path] = d.makeTemplate(e, tplMetricsChainNodeconn)
}

func (d *Dashboard) handleMetricsChainNodeconn(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
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
		Metrics:            metrics,
	})
}

type MetricsChainNodeconnTemplateParams struct {
	BaseTemplateParams
	Metrics nodeconnmetrics.NodeConnectionMessagesMetrics
}
