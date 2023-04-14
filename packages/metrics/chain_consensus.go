package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainConsensusMetrics interface {
	VMRun(duration time.Duration, reqCount int)
}

var (
	_ IChainConsensusMetrics = &emptyChainConsensusMetric{}
	_ IChainConsensusMetrics = &chainConsensusMetric{}
)

type emptyChainConsensusMetric struct{}

func NewEmptyChainConsensusMetric() IChainConsensusMetrics                      { return &emptyChainConsensusMetric{} }
func (m *emptyChainConsensusMetric) VMRun(duration time.Duration, reqCount int) {}

type chainConsensusMetric struct {
	provider      *ChainMetricsProvider
	metricsLabels prometheus.Labels
}

func newChainConsensusMetric(provider *ChainMetricsProvider, chainID isc.ChainID) *chainConsensusMetric {
	metricsLabels := getChainLabels(chainID)

	// init values so they appear in prometheus
	provider.consensusVMRunTime.With(metricsLabels)
	provider.consensusVMRunTimePerReq.With(metricsLabels)
	provider.consensusVMRunReqCount.With(metricsLabels)

	return &chainConsensusMetric{
		provider:      provider,
		metricsLabels: metricsLabels,
	}
}

func (m *chainConsensusMetric) VMRun(duration time.Duration, reqCount int) {
	d := duration.Seconds()
	r := float64(reqCount)
	m.provider.consensusVMRunTime.With(m.metricsLabels).Observe(d)
	m.provider.consensusVMRunTimePerReq.With(m.metricsLabels).Observe(d / r)
	m.provider.consensusVMRunReqCount.With(m.metricsLabels).Observe(r)
}
