package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainSnapshotsMetricsProvider struct {
	created        *countAndMaxMetrics
	createDuration *prometheus.HistogramVec
}

func newChainSnapshotsMetricsProvider() *ChainSnapshotsMetricsProvider {
	return &ChainSnapshotsMetricsProvider{
		created: newCountAndMaxMetrics(
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "iota_wasp",
				Subsystem: "snapshots",
				Name:      "created",
				Help:      "Total number of snapshots created",
			}, []string{labelNameChain}),
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "iota_wasp",
				Subsystem: "snapshots",
				Name:      "max_index",
				Help:      "Largest index of created snapshot",
			}, []string{labelNameChain}),
		),
		createDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "snapshots",
			Name:      "create_duration",
			Help:      "The duration (s) of creating snapshot and storing it in file system",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
	}
}

func (p *ChainSnapshotsMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.createDuration,
	)
	reg.MustRegister(p.created.collectors()...)
}

func (p *ChainSnapshotsMetricsProvider) createForChain(chainID isc.ChainID) *ChainSnapshotsMetrics {
	return newChainSnapshotsMetrics(p, chainID)
}

type ChainSnapshotsMetrics struct {
	labels     prometheus.Labels
	collectors *ChainSnapshotsMetricsProvider
}

func newChainSnapshotsMetrics(collectors *ChainSnapshotsMetricsProvider, chainID isc.ChainID) *ChainSnapshotsMetrics {
	labels := getChainLabels(chainID)

	// init values so they appear in prometheus
	collectors.created.with(labels)
	collectors.createDuration.With(labels)

	return &ChainSnapshotsMetrics{
		collectors: collectors,
		labels:     labels,
	}
}

func (m *ChainSnapshotsMetrics) SnapshotCreated(duration time.Duration, stateIndex uint32) {
	m.collectors.createDuration.With(m.labels).Observe(duration.Seconds())
	m.collectors.created.countValue(m.labels, float64(stateIndex))
}
