package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type IStateManagerMetrics interface {
	SetBlockSize(blockIndex uint32, size float64)
	SetLastSeenStateIndex(stateIndex uint32)
}

type IConsensusMetrics interface {
	SetVMRunTime(time.Duration)
	IncVMRunsCounter()
}

type IMempoolMetrics interface {
	IncRequestsReceived(isc.Request)
	IncRequestsProcessed()
	SetRequestProcessingTime(isc.RequestID, time.Duration)
	IncBlocksPerChain()
}

type IChainMetric interface {
	IStateManagerMetrics
	IConsensusMetrics
	IMempoolMetrics
	IncRequestsAckMessages()
	IncMessagesReceived()
	SetCurrentStateIndex(stateIndex uint32)
}

var (
	_ IChainMetric = &ChainMetric{}
	_ IChainMetric = &emptyChainMetric{}
)

type emptyChainMetric struct{}

func NewEmptyChainMetric() IChainMetric {
	return &emptyChainMetric{}
}

func (m *emptyChainMetric) IncRequestsReceived(_ isc.Request)                         {}
func (m *emptyChainMetric) IncRequestsProcessed()                                     {}
func (m *emptyChainMetric) IncRequestsAckMessages()                                   {}
func (m *emptyChainMetric) SetRequestProcessingTime(_ isc.RequestID, _ time.Duration) {}
func (m *emptyChainMetric) IncMessagesReceived()                                      {}
func (m *emptyChainMetric) SetVMRunTime(_ time.Duration)                              {}
func (m *emptyChainMetric) IncVMRunsCounter()                                         {}
func (m *emptyChainMetric) IncBlocksPerChain()                                        {}
func (m *emptyChainMetric) SetBlockSize(_ uint32, _ float64)                          {}
func (m *emptyChainMetric) SetCurrentStateIndex(stateIndex uint32)                    {}
func (m *emptyChainMetric) SetLastSeenStateIndex(stateIndex uint32)                   {}

type ChainMetric struct {
	chainMetrics          *ChainMetrics
	chainID               isc.ChainID
	lastSeenStateIndexVal uint32
}

func NewChainMetric(chainMetrics *ChainMetrics, chainID isc.ChainID) IChainMetric {
	return &ChainMetric{
		chainMetrics:          chainMetrics,
		chainID:               chainID,
		lastSeenStateIndexVal: 0,
	}
}

func (c *ChainMetric) IncRequestsReceived(req isc.Request) {
	if req.IsOffLedger() {
		c.chainMetrics.requestsReceivedOffLedger.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
	} else {
		c.chainMetrics.requestsReceivedOnLedger.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
	}
}

func (c *ChainMetric) IncRequestsProcessed() {
	c.chainMetrics.requestsProcessed.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *ChainMetric) IncRequestsAckMessages() {
	c.chainMetrics.requestsAckMessages.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *ChainMetric) SetRequestProcessingTime(reqID isc.RequestID, elapse time.Duration) {
	c.chainMetrics.requestsProcessingTime.With(prometheus.Labels{"chain": c.chainID.String(), "request": reqID.String()}).Set(elapse.Seconds())
}

func (c *ChainMetric) IncMessagesReceived() {
	c.chainMetrics.messagesReceived.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *ChainMetric) SetVMRunTime(elapse time.Duration) {
	c.chainMetrics.vmRunTime.With(prometheus.Labels{"chain": c.chainID.String()}).Set(elapse.Seconds())
}

func (c *ChainMetric) IncVMRunsCounter() {
	c.chainMetrics.vmRunsTotal.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *ChainMetric) IncBlocksPerChain() {
	c.chainMetrics.blocksTotalPerChain.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *ChainMetric) SetBlockSize(blockIndex uint32, blockSize float64) {
	c.chainMetrics.blockSizesPerChain.With(prometheus.Labels{"chain": c.chainID.String(), "block_index": fmt.Sprintf("%d", blockIndex)}).Set(blockSize)
}

func (c *ChainMetric) SetCurrentStateIndex(stateIndex uint32) {
	c.chainMetrics.stateIndexCurrent.With(prometheus.Labels{"chain": c.chainID.String()}).Set(float64(stateIndex))
}

func (c *ChainMetric) SetLastSeenStateIndex(stateIndex uint32) {
	if c.lastSeenStateIndexVal >= stateIndex {
		return
	}
	c.lastSeenStateIndexVal = stateIndex
	c.chainMetrics.stateIndexLatestSeen.With(prometheus.Labels{"chain": c.chainID.String()}).Set(float64(stateIndex))
}

// ChainMetrics holds all metrics for all chains per chain
type ChainMetrics struct {
	nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics

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
func NewChainMetrics(nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics) *ChainMetrics {
	return &ChainMetrics{
		nodeConnectionMetrics: nodeConnectionMetrics,

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

func (m *ChainMetrics) PrometheusCollectors() []prometheus.Collector {
	return append(m.nodeConnectionMetrics.PrometheusCollectors(),
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
		m.stateIndexLatestSeen)
}

func (m *ChainMetrics) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return m.nodeConnectionMetrics
}
