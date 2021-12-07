package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed templates/metrics.tmpl
var tplMetrics string

func (d *Dashboard) metricsInit(e *echo.Echo, r renderer) Tab {
	ret := d.initMetrics(e, r)
	d.initMetricsNodeconn(e, r)
	d.initMetricsNodeconnMessages(e, r)
	return ret
}

func (d *Dashboard) initMetrics(e *echo.Echo, r renderer) Tab {
	route := e.GET("/metrics", d.handleMetrics)
	route.Name = "metrics"

	r[route.Path] = d.makeTemplate(e, tplMetrics)

	return Tab{
		Path:  route.Path,
		Title: "Metrics",
		Href:  route.Path,
	}
}

func (d *Dashboard) handleMetrics(c echo.Context) error {
	return c.Render(http.StatusOK, c.Path(), &MetricsTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
	})
}

type MetricsTemplateParams struct {
	BaseTemplateParams
}
