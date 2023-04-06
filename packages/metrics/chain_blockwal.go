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
	provider.blockWALBlocksAdded.With(metricsLabels)
	provider.blockWALMaxBlockIndex.With(metricsLabels)

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
	m.provider.blockWALBlocksAdded.With(m.metricsLabels).Inc()
	blockIndexInt64 := int64(blockIndex)
	if m.maxBlockIndex < 0 {
		m.provider.blockWALMaxBlockIndex.With(m.metricsLabels).Add(float64(blockIndexInt64))
		m.maxBlockIndex = blockIndexInt64
	} else {
		diff := blockIndexInt64 - m.maxBlockIndex
		if diff > 0 {
			m.provider.blockWALMaxBlockIndex.With(m.metricsLabels).Add(float64(diff))
			m.maxBlockIndex = blockIndexInt64
		}
	}
}
