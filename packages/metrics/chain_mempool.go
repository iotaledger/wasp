package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainMempoolMetricsProvider struct {
	blocksTotalPerChain       *prometheus.CounterVec
	requestsReceivedOffLedger *prometheus.CounterVec
	requestsReceivedOnLedger  *prometheus.CounterVec
	requestsProcessed         *prometheus.CounterVec
	requestsAckMessages       *prometheus.CounterVec
	requestsProcessingTime    *prometheus.GaugeVec

	timePoolSize      *prometheus.GaugeVec
	onLedgerPoolSize  *prometheus.GaugeVec
	onLedgerReqTime   *prometheus.HistogramVec
	offLedgerPoolSize *prometheus.GaugeVec
	offLedgerReqTime  *prometheus.HistogramVec
	totalSize         *prometheus.GaugeVec
	missingReqs       *prometheus.GaugeVec
}

func newChainMempoolMetricsProvider() *ChainMempoolMetricsProvider {
	return &ChainMempoolMetricsProvider{
		blocksTotalPerChain: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "blocks",
			Name:      "total",
			Help:      "Number of blocks per chain",
		}, []string{labelNameChain}),
		requestsReceivedOffLedger: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "off_ledger_total",
			Help:      "Number of off-ledger requests made to chain",
		}, []string{labelNameChain}),
		requestsReceivedOnLedger: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "on_ledger_total",
			Help:      "Number of on-ledger requests made to the chain",
		}, []string{labelNameChain}),
		requestsProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "processed_total",
			Help:      "Number of requests processed per chain",
		}, []string{labelNameChain}),
		requestsAckMessages: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "received_acks_total",
			Help:      "Number of received request acknowledgements per chain",
		}, []string{labelNameChain}),
		requestsProcessingTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "processing_time",
			Help:      "Time to process requests per chain",
		}, []string{labelNameChain}),
		timePoolSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "mempool",
			Name:      "time_pool_size",
			Help:      "Number of postponed requests in mempool.",
		}, []string{labelNameChain}),
		onLedgerPoolSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "mempool",
			Name:      "on_ledger_pool_size",
			Help:      "Number of On Ledger requests in mempool.",
		}, []string{labelNameChain}),
		onLedgerReqTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "mempool",
			Name:      "on_ledger_req_time",
			Help:      "Time (s) an on-ledger request stayed in the mempool before removing it.",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		offLedgerPoolSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "mempool",
			Name:      "off_ledger_pool_size",
			Help:      "Number of Off Ledger requests in mempool.",
		}, []string{labelNameChain}),
		offLedgerReqTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "mempool",
			Name:      "off_ledger_req_time",
			Help:      "Time (s) an off-ledger request stayed in the mempool before removing it.",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain}),
		totalSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "mempool",
			Name:      "total_pool_size",
			Help:      "Total requests in mempool.",
		}, []string{labelNameChain}),
		missingReqs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "mempool",
			Name:      "missing_reqs",
			Help:      "Number of requests missing at this node (asking others to send them).",
		}, []string{labelNameChain}),
	}
}

func (p *ChainMempoolMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.blocksTotalPerChain,
		p.requestsReceivedOffLedger,
		p.requestsReceivedOnLedger,
		p.requestsProcessed,
		p.requestsAckMessages,
		p.requestsProcessingTime,
		p.timePoolSize,
		p.onLedgerPoolSize,
		p.onLedgerReqTime,
		p.offLedgerPoolSize,
		p.offLedgerReqTime,
		p.totalSize,
		p.missingReqs,
	)
}

func (p *ChainMempoolMetricsProvider) createForChain(chainID isc.ChainID) *ChainMempoolMetrics {
	return newChainMempoolMetrics(p, chainID)
}

type ChainMempoolMetrics struct {
	collectors *ChainMempoolMetricsProvider
	labels     prometheus.Labels

	vTimePoolSize      int
	vOnLedgerPoolSize  int
	vOffLedgerPoolSize int
}

func newChainMempoolMetrics(collectors *ChainMempoolMetricsProvider, chainID isc.ChainID) *ChainMempoolMetrics {
	labels := getChainLabels(chainID)

	// init values so they appear in prometheus
	collectors.blocksTotalPerChain.With(labels)
	collectors.requestsReceivedOffLedger.With(labels)
	collectors.requestsReceivedOnLedger.With(labels)
	collectors.requestsProcessed.With(labels)
	collectors.requestsAckMessages.With(labels)
	collectors.requestsProcessingTime.With(labels)

	collectors.timePoolSize.With(labels)
	collectors.onLedgerPoolSize.With(labels)
	collectors.onLedgerReqTime.With(labels)
	collectors.offLedgerPoolSize.With(labels)
	collectors.offLedgerReqTime.With(labels)
	collectors.totalSize.With(labels)
	collectors.missingReqs.With(labels)

	return &ChainMempoolMetrics{
		collectors: collectors,
		labels:     labels,
	}
}

func (m *ChainMempoolMetrics) IncBlocksPerChain() {
	m.collectors.blocksTotalPerChain.With(m.labels).Inc()
}

func (m *ChainMempoolMetrics) IncRequestsReceived(request isc.Request) {
	if request.IsOffLedger() {
		m.collectors.requestsReceivedOffLedger.With(m.labels).Inc()
	} else {
		m.collectors.requestsReceivedOnLedger.With(m.labels).Inc()
	}
}

func (m *ChainMempoolMetrics) IncRequestsProcessed() {
	m.collectors.requestsProcessed.With(m.labels).Inc()
}

func (m *ChainMempoolMetrics) SetRequestProcessingTime(duration time.Duration) {
	m.collectors.requestsProcessingTime.With(m.labels).Set(duration.Seconds())
}

func (m *ChainMempoolMetrics) SetTimePoolSize(count int) {
	m.vTimePoolSize = count
	m.collectors.timePoolSize.With(m.labels).Set(float64(count))
	m.deriveTotalPoolSize()
}

func (m *ChainMempoolMetrics) SetOnLedgerPoolSize(count int) {
	m.vOnLedgerPoolSize = count
	m.collectors.onLedgerPoolSize.With(m.labels).Set(float64(count))
	m.deriveTotalPoolSize()
}

func (m *ChainMempoolMetrics) SetOffLedgerPoolSize(count int) {
	m.vOffLedgerPoolSize = count
	m.collectors.offLedgerPoolSize.With(m.labels).Set(float64(count))
	m.deriveTotalPoolSize()
}

func (m *ChainMempoolMetrics) deriveTotalPoolSize() {
	m.collectors.totalSize.With(m.labels).Set(float64(m.vTimePoolSize + m.vOnLedgerPoolSize + m.vOffLedgerPoolSize))
}

func (m *ChainMempoolMetrics) SetMissingReqs(count int) {
	m.collectors.missingReqs.With(m.labels).Set(float64(count))
}

func (m *ChainMempoolMetrics) SetOnLedgerReqTime(d time.Duration) {
	m.collectors.onLedgerReqTime.With(m.labels).Observe(d.Seconds())
}

func (m *ChainMempoolMetrics) SetOffLedgerReqTime(d time.Duration) {
	m.collectors.offLedgerReqTime.With(m.labels).Observe(d.Seconds())
}
