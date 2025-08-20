package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainNodeConnMetricsProvider struct {
	l1RequestReceived *prometheus.CounterVec
	l1AnchorReceived  *prometheus.CounterVec
	txPublishStarted  *prometheus.CounterVec
	txPublishResult   *prometheus.HistogramVec
}

func newChainNodeConnMetricsProvider() *ChainNodeConnMetricsProvider {
	return &ChainNodeConnMetricsProvider{
		l1RequestReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "node_conn",
			Name:      "l1_request_received",
			Help:      "A number of confirmed requests received from L1.",
		}, []string{labelNameChain}),
		l1AnchorReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "node_conn",
			Name:      "l1_anchor_received",
			Help:      "A number of confirmed anchor received from L1.",
		}, []string{labelNameChain}),
		txPublishStarted: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "node_conn",
			Name:      "tx_publish_started",
			Help:      "A number of transactions submitted for publication in L1.",
		}, []string{labelNameChain}),
		txPublishResult: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "node_conn",
			Name:      "tx_publish_result",
			Help:      "The duration (s) to publish a transaction.",
			Buckets:   postTimeBuckets,
		}, []string{labelNameChain, labelTxPublishResult}),
	}
}

func (p *ChainNodeConnMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.l1RequestReceived,
		p.l1AnchorReceived,
		p.txPublishStarted,
		p.txPublishResult,
	)
}

func (p *ChainNodeConnMetricsProvider) createForChain(chainID isc.ChainID) *ChainNodeConnMetrics {
	return newChainNodeConnMetrics(p, chainID)
}

type ChainNodeConnMetrics struct {
	ncL1RequestReceived prometheus.Counter
	ncL1AnchorReceived  prometheus.Counter
	ncTXPublishStarted  prometheus.Counter
	ncTXPublishResult   map[bool]prometheus.Observer
}

func newChainNodeConnMetrics(collectors *ChainNodeConnMetricsProvider, chainID isc.ChainID) *ChainNodeConnMetrics {
	labels := getChainLabels(chainID)
	return &ChainNodeConnMetrics{
		ncL1RequestReceived: collectors.l1RequestReceived.With(labels),
		ncL1AnchorReceived:  collectors.l1AnchorReceived.With(labels),
		ncTXPublishStarted:  collectors.txPublishStarted.With(labels),
		ncTXPublishResult: map[bool]prometheus.Observer{
			true:  collectors.txPublishResult.MustCurryWith(labels).With(prometheus.Labels{labelTxPublishResult: "confirmed"}),
			false: collectors.txPublishResult.MustCurryWith(labels).With(prometheus.Labels{labelTxPublishResult: "rejected"}),
		},
	}
}

func (m *ChainNodeConnMetrics) L1RequestReceived() {
	m.ncL1RequestReceived.Inc()
}

func (m *ChainNodeConnMetrics) L1AnchorReceived() {
	m.ncL1AnchorReceived.Inc()
}

func (m *ChainNodeConnMetrics) TXPublishStarted() {
	m.ncTXPublishStarted.Inc()
}

func (m *ChainNodeConnMetrics) TXPublishResult(confirmed bool, duration time.Duration) {
	m.ncTXPublishResult[confirmed].Observe(duration.Seconds())
}
