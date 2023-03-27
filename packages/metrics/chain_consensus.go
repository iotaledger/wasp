package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainConsensusMetrics interface {
	SetVMRunTime(time.Duration)
	IncVMRunsCounter()
}

var (
	_ IChainConsensusMetrics = &emptyChainConsensusMetric{}
	_ IChainConsensusMetrics = &chainConsensusMetric{}
)

type emptyChainConsensusMetric struct{}

func NewEmptyChainConsensusMetric() IChainConsensusMetrics               { return &emptyChainConsensusMetric{} }
func (m *emptyChainConsensusMetric) SetVMRunTime(duration time.Duration) {}
func (m *emptyChainConsensusMetric) IncVMRunsCounter()                   {}

type chainConsensusMetric struct {
	provider      *ChainMetricsProvider
	metricsLabels prometheus.Labels
}

func newChainConsensusMetric(provider *ChainMetricsProvider, chainID isc.ChainID) *chainConsensusMetric {
	metricsLabels := getChainLabels(chainID)

	// init values so they appear in prometheus
	provider.vmRunTime.With(metricsLabels)
	provider.vmRunsTotal.With(metricsLabels)

	return &chainConsensusMetric{
		provider:      provider,
		metricsLabels: metricsLabels,
	}
}

func (m *chainConsensusMetric) SetVMRunTime(duration time.Duration) {
	m.provider.vmRunTime.With(m.metricsLabels).Set(duration.Seconds())
}

func (m *chainConsensusMetric) IncVMRunsCounter() {
	m.provider.vmRunsTotal.With(m.metricsLabels).Inc()
}
