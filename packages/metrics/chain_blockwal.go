package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainBlockWALMetricsProvider struct {
	failedWrites *prometheus.CounterVec
	failedReads  *prometheus.CounterVec
	blocksAdded  *countAndMaxMetrics
}

func newChainBlockWALMetricsProvider() *ChainBlockWALMetricsProvider {
	return &ChainBlockWALMetricsProvider{
		failedWrites: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "failed_writes_total",
			Help:      "Total number of writes to WAL that failed",
		}, []string{labelNameChain}),
		failedReads: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "failed_reads_total",
			Help:      "Total number of reads failed while replaying WAL",
		}, []string{labelNameChain}),
		blocksAdded: newCountAndMaxMetrics(
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "iota_wasp",
				Subsystem: "wal",
				Name:      "blocks_added",
				Help:      "Total number of blocks added into WAL",
			}, []string{labelNameChain}),
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "iota_wasp",
				Subsystem: "wal",
				Name:      "max_block_index",
				Help:      "Largest index of block added into WAL",
			}, []string{labelNameChain}),
		),
	}
}

func (p *ChainBlockWALMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.failedReads,
		p.failedWrites,
	)
	reg.MustRegister(p.blocksAdded.collectors()...)
}

func (p *ChainBlockWALMetricsProvider) createForChain(chainID isc.ChainID) *ChainBlockWALMetrics {
	return newChainBlockWALMetrics(p, chainID)
}

type ChainBlockWALMetrics struct {
	labels     prometheus.Labels
	collectors *ChainBlockWALMetricsProvider
}

func newChainBlockWALMetrics(collectors *ChainBlockWALMetricsProvider, chainID isc.ChainID) *ChainBlockWALMetrics {
	labels := getChainLabels(chainID)

	// init values so they appear in prometheus
	collectors.failedWrites.With(labels)
	collectors.failedReads.With(labels)
	collectors.blocksAdded.with(labels)

	return &ChainBlockWALMetrics{
		collectors: collectors,
		labels:     labels,
	}
}

func (m *ChainBlockWALMetrics) IncFailedWrites() {
	m.collectors.failedWrites.With(m.labels).Inc()
}

func (m *ChainBlockWALMetrics) IncFailedReads() {
	m.collectors.failedReads.With(m.labels).Inc()
}

func (m *ChainBlockWALMetrics) BlockWritten(blockIndex uint32) {
	m.collectors.blocksAdded.countValue(m.labels, float64(blockIndex))
}
