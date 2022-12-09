package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type Metrics struct {
	nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics
	lastSeenStateIndexVal uint32

	offLedgerRequestCounter *prometheus.CounterVec
	onLedgerRequestCounter  *prometheus.CounterVec
	processedRequestCounter *prometheus.CounterVec
	messagesReceived        *prometheus.CounterVec
	requestAckMessages      *prometheus.CounterVec
	currentStateIndex       *prometheus.GaugeVec
	requestProcessingTime   *prometheus.GaugeVec
	vmRunTime               *prometheus.GaugeVec
	vmRunCounter            *prometheus.CounterVec
	blocksPerChain          *prometheus.CounterVec
	blockSizes              *prometheus.GaugeVec
	lastSeenStateIndex      *prometheus.GaugeVec
}

func New(nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics) *Metrics {
	return &Metrics{
		nodeConnectionMetrics: nodeConnectionMetrics,
		lastSeenStateIndexVal: 0,

		offLedgerRequestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "off_ledger_request_counter",
			Help:      "Number of off-ledger requests made to chain",
		}, []string{"chain"}),

		onLedgerRequestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "on_ledger_request_counter",
			Help:      "Number of on-ledger requests made to the chain",
		}, []string{"chain"}),

		processedRequestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "processed_request_counter",
			Help:      "Number of requests processed",
		}, []string{"chain"}),

		messagesReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "messages_received_per_chain",
			Help:      "Number of messages received",
		}, []string{"chain"}),

		requestAckMessages: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "receive_requests_acknowledgement_message",
			Help:      "Receive request acknowledgement messages per chain",
		}, []string{"chain"}),

		currentStateIndex: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "current_state_index",
			Help:      "The current chain state index.",
		}, []string{"chain"}),

		requestProcessingTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "request_processing_time",
			Help:      "Time to process request",
		}, []string{"chain", "request"}),

		vmRunTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "vm_run_time",
			Help:      "Time it takes to run the vm",
		}, []string{"chain"}),

		vmRunCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "vm_run_counter",
			Help:      "Time it takes to run the vm",
		}, []string{"chain"}),

		blocksPerChain: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "block_counter",
			Help:      "Number of blocks per chain",
		}, []string{"chain"}),

		blockSizes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "block_size",
			Help:      "Block sizes",
		}, []string{"block_index", "chain"}),

		lastSeenStateIndex: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota",
			Subsystem: "wasp_stats",
			Name:      "last_seen_state_index",
			Help:      "Last seen state index",
		}, []string{"chain"}),
	}
}

func (m *Metrics) NewChainMetrics(chainID isc.ChainID) ChainMetrics {
	return &chainMetricsObj{
		metrics: m,
		chainID: chainID,
	}
}

func (m *Metrics) Register(registry *prometheus.Registry) {
	m.nodeConnectionMetrics.Register(registry)

	registry.MustRegister(
		m.offLedgerRequestCounter,
		m.onLedgerRequestCounter,
		m.processedRequestCounter,
		m.messagesReceived,
		m.requestAckMessages,
		m.currentStateIndex,
		m.requestProcessingTime,
		m.vmRunTime,
		m.vmRunCounter,
		m.blocksPerChain,
		m.blockSizes,
		m.lastSeenStateIndex,
	)
}

func (m *Metrics) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return m.nodeConnectionMetrics
}
