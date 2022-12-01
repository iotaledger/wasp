package smGPAUtils

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type blockWALMetrics struct {
	segments     prometheus.Counter
	failedWrites prometheus.Counter
	failedReads  prometheus.Counter
}

var once sync.Once

func newBlockWALMetrics() *blockWALMetrics {
	m := &blockWALMetrics{}

	m.segments = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wasp_wal_total_segments",
		Help: "Total number of segment files",
	})

	m.failedWrites = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wasp_wal_failed_writes",
		Help: "Total number of writes to WAL that failed",
	})

	m.failedReads = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wasp_wal_failed_reads",
		Help: "Total number of reads failed while replaying WAL",
	})

	once.Do(func() {
		prometheus.MustRegister(
			m.segments,
			m.failedWrites,
			m.failedReads,
		)
	})
	return m
}
