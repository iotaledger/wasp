package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type Metrics struct {
	nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics
	lastSeenStateIndexVal uint32

	requestsReceivedOffLedger *prometheus.CounterVec
	requestsReceivedOnLedger  *prometheus.CounterVec
	requestsProcessed         *prometheus.CounterVec
	requestsAckMessages       *prometheus.CounterVec
	requestsProcessingTime    *prometheus.GaugeVec
	messagesReceived          *prometheus.CounterVec
	vmRunTime                 *prometheus.GaugeVec
	vmRunsTotal               *prometheus.CounterVec
	blocksTotalPerChain       *prometheus.CounterVec
	blockSizesPerChain        *prometheus.GaugeVec
	stateIndexCurrent         *prometheus.GaugeVec
	stateIndexLatestSeen      *prometheus.GaugeVec
}

//nolint:funlen
func New(nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics) *Metrics {
	return &Metrics{
		nodeConnectionMetrics: nodeConnectionMetrics,
		lastSeenStateIndexVal: 0,

		//
		// requests
		//
		requestsReceivedOffLedger: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "off_ledger_total",
			Help:      "Number of off-ledger requests made to chain",
		}, []string{"chain"}),

		requestsReceivedOnLedger: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "on_ledger_total",
			Help:      "Number of on-ledger requests made to the chain",
		}, []string{"chain"}),

		requestsProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "processed_total",
			Help:      "Number of requests processed per chain",
		}, []string{"chain"}),

		requestsAckMessages: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "received_acks_total",
			Help:      "Number of received request acknowledgements per chain",
		}, []string{"chain"}),

		requestsProcessingTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "requests",
			Name:      "processing_time",
			Help:      "Time to process requests per chain",
		}, []string{"chain", "request"}),

		//
		// Messages
		//
		messagesReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "messages",
			Name:      "received_total",
			Help:      "Number of messages received per chain",
		}, []string{"chain"}),

		//
		// VM
		//
		vmRunTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "vm",
			Name:      "run_time",
			Help:      "Time it takes to run the vm per chain",
		}, []string{"chain"}),

		vmRunsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "vm",
			Name:      "runs_total",
			Help:      "Number of vm runs per chain",
		}, []string{"chain"}),

		//
		// Blocks
		//
		blocksTotalPerChain: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "blocks",
			Name:      "total",
			Help:      "Number of blocks per chain",
		}, []string{"chain"}),

		blockSizesPerChain: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "blocks",
			Name:      "sizes",
			Help:      "Block sizes per chain",
		}, []string{"block_index", "chain"}),

		//
		// State
		//
		stateIndexCurrent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "index_current",
			Help:      "The current state index per chain",
		}, []string{"chain"}),

		stateIndexLatestSeen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "index_latest_seen",
			Help:      "Latest seen state index per chain",
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
		m.requestsReceivedOffLedger,
		m.requestsReceivedOnLedger,
		m.requestsProcessed,
		m.requestsAckMessages,
		m.requestsProcessingTime,
		m.messagesReceived,
		m.vmRunTime,
		m.vmRunsTotal,
		m.blocksTotalPerChain,
		m.blockSizesPerChain,
		m.stateIndexCurrent,
		m.stateIndexLatestSeen,
	)
}

func (m *Metrics) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return m.nodeConnectionMetrics
}
