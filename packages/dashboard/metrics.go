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
	/*chains, err := d.fetchChains()
	if err != nil {
		return err
	}*/
	return c.Render(http.StatusOK, c.Path(), &MetricsTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
		// Chains:             chains,
	})
}

/*func (d *Dashboard) fetchChains() ([]*ChainOverview, error) {
	crs, err := d.wasp.GetChainRecords()
	if err != nil {
		return nil, err
	}
	r := make([]*ChainOverview, len(crs))
	for i, cr := range crs {
		info, err := d.fetchRootInfo(cr.ChainID)
		r[i] = &ChainOverview{
			ChainRecord: cr,
			RootInfo:    info,
			Error:       err,
		}
	}
	return r, nil
}*/

type MetricsTemplateParams struct {
	BaseTemplateParams
	Chains []*ChainOverview
}
