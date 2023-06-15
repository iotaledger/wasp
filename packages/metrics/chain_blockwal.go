package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainBlockWALMetrics interface {
	IncFailedWrites()
	IncFailedReads()
	BlockWritten(uint32)
}

var (
	_ IChainBlockWALMetrics = &emptyChainBlockWALMetrics{}
	_ IChainBlockWALMetrics = &chainBlockWALMetrics{}
)

type emptyChainBlockWALMetrics struct{}

func NewEmptyChainBlockWALMetrics() IChainBlockWALMetrics { return &emptyChainBlockWALMetrics{} }
func (m *emptyChainBlockWALMetrics) IncFailedWrites()     {}
func (m *emptyChainBlockWALMetrics) IncFailedReads()      {}
func (m *emptyChainBlockWALMetrics) BlockWritten(uint32)  {}

type chainBlockWALMetrics struct {
	provider      *ChainMetricsProvider
	metricsLabels prometheus.Labels

	maxBlockIndex int64
}

func newChainBlockWALMetrics(provider *ChainMetricsProvider, chainID isc.ChainID) *chainBlockWALMetrics {
	metricsLabels := getChainLabels(chainID)

	// init values so they appear in prometheus
	provider.blockWALFailedWrites.With(metricsLabels)
	provider.blockWALFailedReads.With(metricsLabels)
	provider.blockWALBlocksAdded.with(metricsLabels)

	return &chainBlockWALMetrics{
		provider:      provider,
		metricsLabels: metricsLabels,
		maxBlockIndex: -1,
	}
}

func (m *chainBlockWALMetrics) IncFailedWrites() {
	m.provider.blockWALFailedWrites.With(m.metricsLabels).Inc()
}

func (m *chainBlockWALMetrics) IncFailedReads() {
	m.provider.blockWALFailedReads.With(m.metricsLabels).Inc()
}

func (m *chainBlockWALMetrics) BlockWritten(blockIndex uint32) {
	m.provider.blockWALBlocksAdded.countValue(m.metricsLabels, float64(blockIndex))
}
