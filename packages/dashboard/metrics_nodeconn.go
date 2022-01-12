package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/labstack/echo/v4"
)

//go:embed templates/metrics_nodeconn.tmpl
var tplMetricsNodeconn string

func metricsNodeconnBreadcrumb(e *echo.Echo) Tab {
	return Tab{
		Path:  e.Reverse("metricsNodeconn"),
		Title: "Metrics: Connection to L1",
		Href:  e.Reverse("metricsNodeconn"),
	}
}

func (d *Dashboard) initMetricsNodeconn(e *echo.Echo, r renderer) {
	route := e.GET("/metrics/nodeconn", d.handleMetricsNodeconn)
	route.Name = "metricsNodeconn"
	r[route.Path] = d.makeTemplate(e, tplMetricsChainNodeconn, tplMetricsNodeconn)
}

func (d *Dashboard) handleMetricsNodeconn(c echo.Context) error {
	metrics, err := d.wasp.GetNodeConnectionMetrics()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	tab := metricsNodeconnBreadcrumb(c.Echo())
	return c.Render(http.StatusOK, c.Path(), &MetricsNodeconnTemplateParams{
		BaseTemplateParams: d.BaseParams(c, tab),
		Metrics:            metrics,
	})
}

type MetricsNodeconnTemplateParams struct {
	BaseTemplateParams
	Metrics nodeconnmetrics.NodeConnectionMetrics
}
