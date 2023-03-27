package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainBlockWALMetrics interface {
	IncFailedWrites()
	IncFailedReads()
	IncSegments()
}

var (
	_ IChainBlockWALMetrics = &emptyChainBlockWALMetrics{}
	_ IChainBlockWALMetrics = &chainBlockWALMetrics{}
)

type emptyChainBlockWALMetrics struct{}

func NewEmptyChainBlockWALMetrics() IChainBlockWALMetrics { return &emptyChainBlockWALMetrics{} }
func (m *emptyChainBlockWALMetrics) IncFailedWrites()     {}
func (m *emptyChainBlockWALMetrics) IncFailedReads()      {}
func (m *emptyChainBlockWALMetrics) IncSegments()         {}

type chainBlockWALMetrics struct {
	provider      *ChainMetricsProvider
	metricsLabels prometheus.Labels
}

func newChainBlockWALMetrics(provider *ChainMetricsProvider, chainID isc.ChainID) *chainBlockWALMetrics {
	metricsLabels := getChainLabels(chainID)

	// init values so they appear in prometheus
	provider.blockWALFailedWrites.With(metricsLabels)
	provider.blockWALFailedReads.With(metricsLabels)
	provider.blockWALSegments.With(metricsLabels)

	return &chainBlockWALMetrics{
		provider:      provider,
		metricsLabels: metricsLabels,
	}
}

func (m *chainBlockWALMetrics) IncFailedWrites() {
	m.provider.blockWALFailedWrites.With(m.metricsLabels).Inc()
}

func (m *chainBlockWALMetrics) IncFailedReads() {
	m.provider.blockWALFailedReads.With(m.metricsLabels).Inc()
}

func (m *chainBlockWALMetrics) IncSegments() {
	m.provider.blockWALSegments.With(m.metricsLabels).Inc()
}
