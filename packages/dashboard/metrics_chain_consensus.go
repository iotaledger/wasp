package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/labstack/echo/v4"
)

//go:embed templates/metrics_chain_consensus.tmpl
var tplMetricsChainConsensus string

func metricsChainConsensusBreadcrumb(e *echo.Echo, chainID *iscp.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("metricsChainConsensus"),
		Title: fmt.Sprintf("Metrics: %.8s: Consensus", chainID.Base58()),
		Href:  e.Reverse("metricsChainConsensus", chainID.Base58()),
	}
}

func (d *Dashboard) initMetricsChainConsensus(e *echo.Echo, r renderer) {
	route := e.GET("/metrics/:chainid/consensus", d.handleMetricsChainConsensus)
	route.Name = "metricsChainConsensus"
	r[route.Path] = d.makeTemplate(e, tplMetricsChainConsensus)
}

func (d *Dashboard) handleMetricsChainConsensus(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	tab := metricsChainConsensusBreadcrumb(c.Echo(), chainID)
	status, err := d.wasp.GetChainConsensusWorkflowStatus(chainID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	pipeMetrics, err := d.wasp.GetChainConsensusPipeMetrics(chainID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	return c.Render(http.StatusOK, c.Path(), &MetricsChainConsensusTemplateParams{
		BaseTemplateParams: d.BaseParams(c, metricsChainBreadcrumb(c.Echo(), chainID), tab),
		ChainID:            chainID.Base58(),
		Status:             status,
		PipeMetrics:        pipeMetrics,
	})
}

type MetricsChainConsensusTemplateParams struct {
	BaseTemplateParams
	ChainID     string
	Status      chain.ConsensusWorkflowStatus
	PipeMetrics chain.ConsensusPipeMetrics
}
