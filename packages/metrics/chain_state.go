package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainStateMetrics interface {
	SetBlockSize(size float64)
	SetCurrentStateIndex(stateIndex uint32)
	SetLastSeenStateIndex(stateIndex uint32)
}

var (
	_ IChainStateMetrics = &emptyChainStateMetric{}
	_ IChainStateMetrics = &chainStateMetric{}
)

type emptyChainStateMetric struct{}

func NewEmptyChainStateMetric() IChainStateMetrics {
	return &emptyChainStateMetric{}
}
func (m *emptyChainStateMetric) SetBlockSize(size float64)               {}
func (m *emptyChainStateMetric) SetCurrentStateIndex(stateIndex uint32)  {}
func (m *emptyChainStateMetric) SetLastSeenStateIndex(stateIndex uint32) {}

type chainStateMetric struct {
	provider              *ChainMetricsProvider
	metricsLabels         prometheus.Labels
	lastSeenStateIndexVal uint32
}

func newChainStateMetric(provider *ChainMetricsProvider, chainID isc.ChainID) *chainStateMetric {
	metricsLabels := getChainLabels(chainID)

	// init values so they appear in prometheus
	provider.blockSizesPerChain.With(metricsLabels)
	provider.stateIndexCurrent.With(metricsLabels)
	provider.stateIndexLatestSeen.With(metricsLabels)

	return &chainStateMetric{
		provider:              provider,
		metricsLabels:         metricsLabels,
		lastSeenStateIndexVal: 0,
	}
}

func (m *chainStateMetric) SetBlockSize(blockSize float64) {
	m.provider.blockSizesPerChain.With(m.metricsLabels).Set(blockSize)
}

func (m *chainStateMetric) SetCurrentStateIndex(stateIndex uint32) {
	m.provider.stateIndexCurrent.With(m.metricsLabels).Set(float64(stateIndex))
}

func (m *chainStateMetric) SetLastSeenStateIndex(stateIndex uint32) {
	if m.lastSeenStateIndexVal >= stateIndex {
		return
	}
	m.lastSeenStateIndexVal = stateIndex
	m.provider.stateIndexLatestSeen.With(m.metricsLabels).Set(float64(stateIndex))
}
