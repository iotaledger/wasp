package dashboard

import (
	_ "embed"
	"net/http"

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
	r[route.Path] = d.makeTemplate(e, tplMetricsNodeconn)
}

func (d *Dashboard) handleMetricsNodeconn(c echo.Context) error {
	chains, err := d.fetchChains()
	if err != nil {
		return err
	}
	tab := metricsNodeconnBreadcrumb(c.Echo())
	return c.Render(http.StatusOK, c.Path(), &MetricsNodeconnTemplateParams{
		BaseTemplateParams: d.BaseParams(c, tab),
		Chains:             chains,
	})
}

type MetricsNodeconnTemplateParams struct {
	BaseTemplateParams
	Chains []*ChainOverview
}
