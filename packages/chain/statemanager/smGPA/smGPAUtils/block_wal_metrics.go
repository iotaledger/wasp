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
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "segment_files_total",
			Help:      "Total number of segment files",
		}),
		failedWrites: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "failed_writes_total",
			Help:      "Total number of writes to WAL that failed",
		}),
		failedReads: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "failed_reads_total",
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
