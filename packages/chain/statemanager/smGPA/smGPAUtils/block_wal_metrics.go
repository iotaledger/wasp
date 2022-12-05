package smGPAUtils

import (
	"github.com/prometheus/client_golang/prometheus"
)

type BlockWALMetrics struct {
	segments     prometheus.Counter
	failedWrites prometheus.Counter
	failedReads  prometheus.Counter
}

func NewBlockWALMetrics() *BlockWALMetrics {
	return &BlockWALMetrics{
		segments: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_wal",
			Name:      "total_segments",
			Help:      "Total number of segment files",
		}),
		failedWrites: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_wal",
			Name:      "failed_writes",
			Help:      "Total number of writes to WAL that failed",
		}),
		failedReads: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_wal",
			Name:      "failed_reads",
			Help:      "Total number of reads failed while replaying WAL",
		}),
	}
}

func (m *BlockWALMetrics) Register(registry *prometheus.Registry) {
	prometheus.MustRegister(
		m.segments,
		m.failedWrites,
		m.failedReads,
	)
}
