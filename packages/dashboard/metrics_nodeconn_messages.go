package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/labstack/echo/v4"
)

//go:embed templates/metrics_nodeconn_messages.tmpl
var tplMetricsNodeconnMessages string

func metricsNodeconnMessagesBreadcrumb(e *echo.Echo, chainID *iscp.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("metricsNodeconnMessages"),
		Title: fmt.Sprintf("Metrics: Connection to L1: %.8s", chainID.Base58()),
		Href:  e.Reverse("metricsNodeconnMessages", chainID.Base58()),
	}
}

func (d *Dashboard) initMetricsNodeconnMessages(e *echo.Echo, r renderer) {
	route := e.GET("/metrics/nodeconn/:chainid", d.handleMetricsNodeconnMessages)
	route.Name = "metricsNodeconnMessages"
	r[route.Path] = d.makeTemplate(e, tplMetricsNodeconnMessages)
}

func (d *Dashboard) handleMetricsNodeconnMessages(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	tab := metricsNodeconnMessagesBreadcrumb(c.Echo(), chainID)
	metrics, err := d.wasp.GetChainNodeConnectionMetrics(chainID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	return c.Render(http.StatusOK, c.Path(), &MetricsNodeconnMessagesTemplateParams{
		BaseTemplateParams: d.BaseParams(c, metricsNodeconnBreadcrumb(c.Echo()), tab),
		Metrics:            metrics,
	})
}

type MetricsNodeconnMessagesTemplateParams struct {
	BaseTemplateParams
	Metrics nodeconnmetrics.NodeConnectionMessagesMetrics
}
