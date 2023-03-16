package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IBlockWALMetric interface {
	IncFailedWrites()
	IncFailedReads()
	IncSegments()
}

var (
	_ IBlockWALMetric = &BlockWALMetric{}
	_ IBlockWALMetric = &emptyBlockWALMetric{}
)

type emptyBlockWALMetric struct{}

func NewEmptyBlockWALMetric() IBlockWALMetric {
	return &emptyBlockWALMetric{}
}

func (m *emptyBlockWALMetric) IncFailedWrites() {}
func (m *emptyBlockWALMetric) IncFailedReads()  {}
func (m *emptyBlockWALMetric) IncSegments()     {}

type BlockWALMetric struct {
	blockWALMetrics *BlockWALMetrics
	chainID         isc.ChainID
}

func NewBlockWALMetric(blockWALMetrics *BlockWALMetrics, chainID isc.ChainID) IBlockWALMetric {
	return &BlockWALMetric{
		blockWALMetrics: blockWALMetrics,
		chainID:         chainID,
	}
}

func (m *BlockWALMetric) IncFailedWrites() {
	m.blockWALMetrics.failedWrites.With(prometheus.Labels{"chain": m.chainID.String()}).Inc()
}

func (m *BlockWALMetric) IncFailedReads() {
	m.blockWALMetrics.failedReads.With(prometheus.Labels{"chain": m.chainID.String()}).Inc()
}

func (m *BlockWALMetric) IncSegments() {
	m.blockWALMetrics.segments.With(prometheus.Labels{"chain": m.chainID.String()}).Inc()
}

type BlockWALMetrics struct {
	failedWrites *prometheus.CounterVec
	failedReads  *prometheus.CounterVec
	segments     *prometheus.CounterVec
}

func NewBlockWALMetrics() *BlockWALMetrics {
	return &BlockWALMetrics{
		segments: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "segment_files_total",
			Help:      "Total number of segment files",
		}, []string{"chain"}),
		failedWrites: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "failed_writes_total",
			Help:      "Total number of writes to WAL that failed",
		}, []string{"chain"}),
		failedReads: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "failed_reads_total",
			Help:      "Total number of reads failed while replaying WAL",
		}, []string{"chain"}),
	}
}

func (m *BlockWALMetrics) PrometheusCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.segments,
		m.failedWrites,
		m.failedReads,
	}
}
