package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type ChainSnapshotsMetricsProvider struct {
	created        *countAndMaxMetrics
	updateDuration *prometheus.HistogramVec
	createDuration *prometheus.HistogramVec
	loadDuration   *prometheus.HistogramVec
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
		updateDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "snapshots",
			Name:      "update_duration",
			Help:      "The duration (s) of updating available snashots list",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		createDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "snapshots",
			Name:      "create_duration",
			Help:      "The duration (s) of creating snapshot and storing it in file system",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		loadDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "snapshots",
			Name:      "load_duration",
			Help:      "The duration (s) of (down)loading snapshot an writing it into the store",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
	}
}

func (p *ChainSnapshotsMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.updateDuration,
		p.createDuration,
		p.loadDuration,
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
	collectors.updateDuration.With(labels)
	collectors.createDuration.With(labels)
	collectors.loadDuration.With(labels)

	return &ChainSnapshotsMetrics{
		collectors: collectors,
		labels:     labels,
	}
}

func (m *ChainSnapshotsMetrics) SnapshotsUpdated(duration time.Duration) {
	m.collectors.updateDuration.With(m.labels).Observe(duration.Seconds())
}

func (m *ChainSnapshotsMetrics) SnapshotCreated(duration time.Duration, stateIndex uint32) {
	m.collectors.createDuration.With(m.labels).Observe(duration.Seconds())
	m.collectors.created.countValue(m.labels, float64(stateIndex))
}

func (m *ChainSnapshotsMetrics) SnapshotLoaded(duration time.Duration) {
	m.collectors.loadDuration.With(m.labels).Observe(duration.Seconds())
}
