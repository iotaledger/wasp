package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainConsensusMetricsProvider struct {
	vmRunTime       *prometheus.HistogramVec
	vmRunTimePerReq *prometheus.HistogramVec
	vmRunReqCount   *prometheus.HistogramVec
}

func newChainConsensusMetricsProvider() *ChainConsensusMetricsProvider {
	return &ChainConsensusMetricsProvider{
		vmRunTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "consensus",
			Name:      "vm_run_time",
			Help:      "Time (s) it takes to run the VM per chain block.",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		vmRunTimePerReq: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "consensus",
			Name:      "vm_run_time_per_req",
			Help:      "Time (s) it takes to run the VM per request.",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		vmRunReqCount: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "consensus",
			Name:      "vm_run_req_count",
			Help:      "Number of requests processed per VM run.",
			Buckets:   recCountBuckets,
		}, []string{labelNameChain}),
	}
}

func (p *ChainConsensusMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.vmRunTime,
		p.vmRunTimePerReq,
		p.vmRunReqCount,
	)
}

func (p *ChainConsensusMetricsProvider) createForChain(chainID isc.ChainID) *ChainConsensusMetrics {
	return newChainConsensusMetrics(p, chainID)
}

type ChainConsensusMetrics struct {
	labels     prometheus.Labels
	collectors *ChainConsensusMetricsProvider
}

func newChainConsensusMetrics(collectors *ChainConsensusMetricsProvider, chainID isc.ChainID) *ChainConsensusMetrics {
	labels := getChainLabels(chainID)

	// init values so they appear in prometheus
	collectors.vmRunTime.With(labels)
	collectors.vmRunTimePerReq.With(labels)
	collectors.vmRunReqCount.With(labels)

	return &ChainConsensusMetrics{
		collectors: collectors,
		labels:     labels,
	}
}

func (m *ChainConsensusMetrics) VMRun(duration time.Duration, reqCount int) {
	d := duration.Seconds()
	r := float64(reqCount)
	m.collectors.vmRunTime.With(m.labels).Observe(d)
	m.collectors.vmRunTimePerReq.With(m.labels).Observe(d / r)
	m.collectors.vmRunReqCount.With(m.labels).Observe(r)
}
