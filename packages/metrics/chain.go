package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
)

const (
	labelNameChain                                  = "chain"
	labelNameMessageType                            = "message_type"
	labelNameInMilestone                            = "in_milestone"
	labelNameInStateOutputMetrics                   = "in_state_output"
	labelNameInAliasOutputMetrics                   = "in_alias_output"
	labelNameInOutputMetrics                        = "in_output"
	labelNameInOnLedgerRequestMetrics               = "in_on_ledger_request"
	labelNameInTxInclusionStateMetrics              = "in_tx_inclusion_state"
	labelNameOutPublishStateTransactionMetrics      = "out_publish_state_transaction"
	labelNameOutPublishGovernanceTransactionMetrics = "out_publish_gov_transaction"
	labelNameOutPullLatestOutputMetrics             = "out_pull_latest_output"
	labelNameOutPullTxInclusionStateMetrics         = "out_pull_tx_inclusion_state"
	labelNameOutPullOutputByIDMetrics               = "out_pull_output_by_id"
)

type IChainMetrics interface {
	IChainBlockWALMetrics
	IChainConsensusMetrics
	IChainMempoolMetrics
	IChainMessageMetrics
	IChainStateMetrics
}

var (
	_ IChainMetrics = &emptyChainMetrics{}
	_ IChainMetrics = &chainMetrics{}
)

type messageMetric[T any] struct {
	provider        *ChainMetricsProvider
	metricsLabels   prometheus.Labels
	messagesCount   atomic.Uint32
	lastMessage     T
	lastMessageTime time.Time
}

func newMessageMetric[T any](provider *ChainMetricsProvider, msgType string) *messageMetric[T] {
	metricsLabels := prometheus.Labels{
		labelNameMessageType: msgType,
	}

	// init values so they appear in prometheus
	provider.messagesL1.With(metricsLabels)
	provider.lastL1MessageTime.With(metricsLabels)

	return &messageMetric[T]{
		provider:      provider,
		metricsLabels: metricsLabels,
	}
}

func (m *messageMetric[T]) IncMessages(msg T, ts ...time.Time) {
	timestamp := time.Now()
	if len(ts) > 0 {
		timestamp = ts[0]
	}

	m.messagesCount.Add(1)
	m.lastMessage = msg
	m.lastMessageTime = timestamp

	m.provider.messagesL1.With(m.metricsLabels).Inc()
	m.provider.lastL1MessageTime.With(m.metricsLabels).Set(float64(timestamp.Unix()))
}

func (m *messageMetric[T]) MessagesTotal() uint32 {
	return m.messagesCount.Load()
}

func (m *messageMetric[T]) LastMessageTime() time.Time {
	return m.lastMessageTime
}

func (m *messageMetric[T]) LastMessage() T {
	return m.lastMessage
}

type emptyChainMetrics struct {
	IChainBlockWALMetrics
	IChainConsensusMetrics
	IChainMempoolMetrics
	IChainMessageMetrics
	IChainStateMetrics
}

func NewEmptyChainMetrics() IChainMetrics {
	return &emptyChainMetrics{
		IChainBlockWALMetrics:  NewEmptyChainBlockWALMetrics(),
		IChainConsensusMetrics: NewEmptyChainConsensusMetric(),
		IChainMempoolMetrics:   NewEmptyChainMempoolMetric(),
		IChainMessageMetrics:   NewEmptyChainMessageMetrics(),
		IChainStateMetrics:     NewEmptyChainStateMetric(),
	}
}

type chainMetrics struct {
	*chainBlockWALMetrics
	*chainConsensusMetric
	*chainMempoolMetric
	*chainMessageMetrics
	*chainStateMetric
}

func newChainMetrics(provider *ChainMetricsProvider, chainID isc.ChainID) *chainMetrics {
	return &chainMetrics{
		chainBlockWALMetrics: newChainBlockWALMetrics(provider, chainID),
		chainConsensusMetric: newChainConsensusMetric(provider, chainID),
		chainMempoolMetric:   newChainMempoolMetric(provider, chainID),
		chainMessageMetrics:  newChainMessageMetrics(provider, chainID),
		chainStateMetric:     newChainStateMetric(provider, chainID),
	}
}

// ChainMetricsProvider holds all metrics for all chains per chain
type ChainMetricsProvider struct {
	// blockWAL
	blockWALFailedWrites *prometheus.CounterVec
	blockWALFailedReads  *prometheus.CounterVec
	blockWALSegments     *prometheus.CounterVec

	// consensus
	vmRunTime   *prometheus.GaugeVec
	vmRunsTotal *prometheus.CounterVec

	// mempool
	blocksTotalPerChain       *prometheus.CounterVec
	requestsReceivedOffLedger *prometheus.CounterVec
	requestsReceivedOnLedger  *prometheus.CounterVec
	requestsProcessed         *prometheus.CounterVec
	requestsAckMessages       *prometheus.CounterVec
	requestsProcessingTime    *prometheus.GaugeVec

	// messages
	chainsRegistered       []isc.ChainID
	messagesL1             *prometheus.CounterVec
	lastL1MessageTime      *prometheus.GaugeVec
	messagesL1Chain        *prometheus.CounterVec
	lastL1MessageTimeChain *prometheus.GaugeVec

	inMilestoneMetrics                     *messageMetric[*nodeclient.MilestoneInfo]
	inStateOutputMetrics                   *messageMetric[*InStateOutput]
	inAliasOutputMetrics                   *messageMetric[*iotago.AliasOutput]
	inOutputMetrics                        *messageMetric[*InOutput]
	inOnLedgerRequestMetrics               *messageMetric[isc.OnLedgerRequest]
	inTxInclusionStateMetrics              *messageMetric[*TxInclusionStateMsg]
	outPublishStateTransactionMetrics      *messageMetric[*StateTransaction]
	outPublishGovernanceTransactionMetrics *messageMetric[*iotago.Transaction]
	outPullLatestOutputMetrics             *messageMetric[interface{}]
	outPullTxInclusionStateMetrics         *messageMetric[iotago.TransactionID]
	outPullOutputByIDMetrics               *messageMetric[iotago.OutputID]

	// state
	blockSizesPerChain   *prometheus.GaugeVec
	stateIndexCurrent    *prometheus.GaugeVec
	stateIndexLatestSeen *prometheus.GaugeVec
}

//nolint:funlen
func NewChainMetricsProvider() *ChainMetricsProvider {
	m := &ChainMetricsProvider{
		//
		// blockWAL
		//
		blockWALFailedWrites: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "failed_writes_total",
			Help:      "Total number of writes to WAL that failed",
		}, []string{labelNameChain}),
		blockWALFailedReads: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "failed_reads_total",
			Help:      "Total number of reads failed while replaying WAL",
		}, []string{labelNameChain}),
		blockWALSegments: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "wal",
			Name:      "segment_files_total",
			Help:      "Total number of segment files",
		}, []string{labelNameChain}),

		//
		// consensus
		//
		vmRunTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "vm",
			Name:      "run_time",
			Help:      "Time it takes to run the vm per chain",
		}, []string{labelNameChain}),

		vmRunsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "vm",
			Name:      "runs_total",
			Help:      "Number of vm runs per chain",
		}, []string{labelNameChain}),

		//
		// mempool
		//
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

		//
		// messages
		//
		chainsRegistered: make([]isc.ChainID, 0),

		messagesL1: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "l1",
			Name:      "messages_total",
			Help:      "Number of messages sent/received by L1 connection",
		}, []string{labelNameMessageType}),

		lastL1MessageTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "l1",
			Name:      "last_message_time",
			Help:      "Last time when a message was sent/received by L1 connection",
		}, []string{labelNameMessageType}),

		messagesL1Chain: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "l1",
			Name:      "messages_total",
			Help:      "Number of messages sent/received by L1 connection of the chain",
		}, []string{labelNameChain, labelNameMessageType}),

		lastL1MessageTimeChain: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "l1",
			Name:      "last_message_time",
			Help:      "Last time when a message was sent/received by L1 connection of the chain",
		}, []string{labelNameChain, labelNameMessageType}),

		//
		// state
		//
		blockSizesPerChain: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "blocks",
			Name:      "sizes",
			Help:      "Block sizes per chain",
		}, []string{labelNameChain}),

		stateIndexCurrent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "index_current",
			Help:      "The current state index per chain",
		}, []string{labelNameChain}),

		stateIndexLatestSeen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "state",
			Name:      "index_latest_seen",
			Help:      "Latest seen state index per chain",
		}, []string{labelNameChain}),
	}

	m.inMilestoneMetrics = newMessageMetric[*nodeclient.MilestoneInfo](m, labelNameInMilestone)
	m.inStateOutputMetrics = newMessageMetric[*InStateOutput](m, labelNameInStateOutputMetrics)
	m.inAliasOutputMetrics = newMessageMetric[*iotago.AliasOutput](m, labelNameInAliasOutputMetrics)
	m.inOutputMetrics = newMessageMetric[*InOutput](m, labelNameInOutputMetrics)
	m.inOnLedgerRequestMetrics = newMessageMetric[isc.OnLedgerRequest](m, labelNameInOnLedgerRequestMetrics)
	m.inTxInclusionStateMetrics = newMessageMetric[*TxInclusionStateMsg](m, labelNameInTxInclusionStateMetrics)
	m.outPublishStateTransactionMetrics = newMessageMetric[*StateTransaction](m, labelNameOutPublishStateTransactionMetrics)
	m.outPublishGovernanceTransactionMetrics = newMessageMetric[*iotago.Transaction](m, labelNameOutPublishGovernanceTransactionMetrics)
	m.outPullLatestOutputMetrics = newMessageMetric[interface{}](m, labelNameOutPullLatestOutputMetrics)
	m.outPullTxInclusionStateMetrics = newMessageMetric[iotago.TransactionID](m, labelNameOutPullTxInclusionStateMetrics)
	m.outPullOutputByIDMetrics = newMessageMetric[iotago.OutputID](m, labelNameOutPullOutputByIDMetrics)

	return m
}

func (m *ChainMetricsProvider) NewChainMetrics(chainID isc.ChainID) IChainMetrics {
	return newChainMetrics(m, chainID)
}

func (m *ChainMetricsProvider) PrometheusCollectorsBlockWAL() []prometheus.Collector {
	return []prometheus.Collector{
		m.blockWALFailedWrites,
		m.blockWALFailedReads,
		m.blockWALSegments,
	}
}

func (m *ChainMetricsProvider) PrometheusCollectorsConsensus() []prometheus.Collector {
	return []prometheus.Collector{
		m.vmRunTime,
		m.vmRunsTotal,
	}
}

func (m *ChainMetricsProvider) PrometheusCollectorsMempool() []prometheus.Collector {
	return []prometheus.Collector{
		m.blocksTotalPerChain,
		m.requestsReceivedOffLedger,
		m.requestsReceivedOnLedger,
		m.requestsProcessed,
		m.requestsAckMessages,
		m.requestsProcessingTime,
	}
}

func (m *ChainMetricsProvider) PrometheusCollectorsChainMessages() []prometheus.Collector {
	return []prometheus.Collector{
		m.messagesL1,
		m.lastL1MessageTime,
		m.messagesL1Chain,
		m.lastL1MessageTimeChain,
	}
}

func (m *ChainMetricsProvider) PrometheusCollectorsChainState() []prometheus.Collector {
	return []prometheus.Collector{
		m.blockSizesPerChain,
		m.stateIndexCurrent,
		m.stateIndexLatestSeen,
	}
}

func (m *ChainMetricsProvider) RegisterChain(chainID isc.ChainID) {
	m.chainsRegistered = append(m.chainsRegistered, chainID)
}

func (m *ChainMetricsProvider) UnregisterChain(chainID isc.ChainID) {
	for i := 0; i < len(m.chainsRegistered); i++ {
		if m.chainsRegistered[i] == chainID {
			// remove the found chain from the slice and return
			m.chainsRegistered = append(m.chainsRegistered[:i], m.chainsRegistered[i+1:]...)
			return
		}
	}
}

func (m *ChainMetricsProvider) RegisteredChains() []isc.ChainID {
	return m.chainsRegistered
}

func (m *ChainMetricsProvider) InMilestone() IMessageMetric[*nodeclient.MilestoneInfo] {
	return m.inMilestoneMetrics
}

func (m *ChainMetricsProvider) InStateOutput() IMessageMetric[*InStateOutput] {
	return m.inStateOutputMetrics
}

func (m *ChainMetricsProvider) InAliasOutput() IMessageMetric[*iotago.AliasOutput] {
	return m.inAliasOutputMetrics
}

func (m *ChainMetricsProvider) InOutput() IMessageMetric[*InOutput] {
	return m.inOutputMetrics
}

func (m *ChainMetricsProvider) InOnLedgerRequest() IMessageMetric[isc.OnLedgerRequest] {
	return m.inOnLedgerRequestMetrics
}

func (m *ChainMetricsProvider) InTxInclusionState() IMessageMetric[*TxInclusionStateMsg] {
	return m.inTxInclusionStateMetrics
}

func (m *ChainMetricsProvider) OutPublishStateTransaction() IMessageMetric[*StateTransaction] {
	return m.outPublishStateTransactionMetrics
}

func (m *ChainMetricsProvider) OutPublishGovernanceTransaction() IMessageMetric[*iotago.Transaction] {
	return m.outPublishGovernanceTransactionMetrics
}

func (m *ChainMetricsProvider) OutPullLatestOutput() IMessageMetric[interface{}] {
	return m.outPullLatestOutputMetrics
}

func (m *ChainMetricsProvider) OutPullTxInclusionState() IMessageMetric[iotago.TransactionID] {
	return m.outPullTxInclusionStateMetrics
}

func (m *ChainMetricsProvider) OutPullOutputByID() IMessageMetric[iotago.OutputID] {
	return m.outPullOutputByIDMetrics
}
