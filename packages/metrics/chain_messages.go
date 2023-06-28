package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
)

type IMessageMetric[T any] interface {
	IncMessages(msg T, ts ...time.Time)
	MessagesTotal() uint32
	LastMessageTime() time.Time
	LastMessage() T
}

// TODO: Review, if they are used/needed.
type ChainMessageMetricsProvider struct {
	messagesL1             *prometheus.CounterVec // TODO: Outdated and should be removed?
	lastL1MessageTime      *prometheus.GaugeVec   // TODO: Outdated and should be removed?
	messagesL1Chain        *prometheus.CounterVec // TODO: Outdated and should be removed?
	lastL1MessageTimeChain *prometheus.GaugeVec   // TODO: Outdated and should be removed?

	inMilestone                     *MessageMetric[*nodeclient.MilestoneInfo] // TODO: Outdated and should be removed?
	inStateOutput                   *MessageMetric[*InStateOutput]            // TODO: Outdated and should be removed?
	inAliasOutput                   *MessageMetric[*iotago.AliasOutput]       // TODO: Outdated and should be removed?
	inOutput                        *MessageMetric[*InOutput]                 // TODO: Outdated and should be removed?
	inOnLedgerRequest               *MessageMetric[isc.OnLedgerRequest]       // TODO: Outdated and should be removed?
	inTxInclusionState              *MessageMetric[*TxInclusionStateMsg]      // TODO: Outdated and should be removed?
	outPublishStateTransaction      *MessageMetric[*StateTransaction]         // TODO: Outdated and should be removed?
	outPublishGovernanceTransaction *MessageMetric[*iotago.Transaction]       // TODO: Outdated and should be removed?
	outPullLatestOutput             *MessageMetric[interface{}]               // TODO: Outdated and should be removed?
	outPullTxInclusionState         *MessageMetric[iotago.TransactionID]      // TODO: Outdated and should be removed?
	outPullOutputByID               *MessageMetric[iotago.OutputID]           // TODO: Outdated and should be removed?
}

func newChainMessageMetricsProvider() *ChainMessageMetricsProvider {
	p := &ChainMessageMetricsProvider{
		messagesL1: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "messages",
			Name:      "messages_total",
			Help:      "Number of messages sent/received by L1 connection",
		}, []string{labelNameMessageType}),
		lastL1MessageTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "messages",
			Name:      "last_message_time",
			Help:      "Last time when a message was sent/received by L1 connection",
		}, []string{labelNameMessageType}),
		messagesL1Chain: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "messages",
			Name:      "chain_messages_total",
			Help:      "Number of messages sent/received by L1 connection of the chain",
		}, []string{labelNameChain, labelNameMessageType}),
		lastL1MessageTimeChain: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "messages",
			Name:      "chain_last_message_time",
			Help:      "Last time when a message was sent/received by L1 connection of the chain",
		}, []string{labelNameChain, labelNameMessageType}),
	}

	p.inMilestone = newMessageMetric[*nodeclient.MilestoneInfo](p, labelNameInMilestone)
	p.inStateOutput = newMessageMetric[*InStateOutput](p, labelNameInStateOutputMetrics)
	p.inAliasOutput = newMessageMetric[*iotago.AliasOutput](p, labelNameInAliasOutputMetrics)
	p.inOutput = newMessageMetric[*InOutput](p, labelNameInOutputMetrics)
	p.inOnLedgerRequest = newMessageMetric[isc.OnLedgerRequest](p, labelNameInOnLedgerRequestMetrics)
	p.inTxInclusionState = newMessageMetric[*TxInclusionStateMsg](p, labelNameInTxInclusionStateMetrics)
	p.outPublishStateTransaction = newMessageMetric[*StateTransaction](p, labelNameOutPublishStateTransactionMetrics)
	p.outPublishGovernanceTransaction = newMessageMetric[*iotago.Transaction](p, labelNameOutPublishGovernanceTransactionMetrics)
	p.outPullLatestOutput = newMessageMetric[interface{}](p, labelNameOutPullLatestOutputMetrics)
	p.outPullTxInclusionState = newMessageMetric[iotago.TransactionID](p, labelNameOutPullTxInclusionStateMetrics)
	p.outPullOutputByID = newMessageMetric[iotago.OutputID](p, labelNameOutPullOutputByIDMetrics)

	return p
}

func (p *ChainMessageMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.messagesL1,
		p.lastL1MessageTime,
		p.messagesL1Chain,
		p.lastL1MessageTimeChain,
	)
}

func (p *ChainMessageMetricsProvider) createForChain(chainID isc.ChainID) *ChainMessageMetrics {
	return &ChainMessageMetrics{
		inStateOutput:                   createChainMessageMetric(p, chainID, labelNameInStateOutputMetrics, p.inStateOutput),
		inAliasOutput:                   createChainMessageMetric(p, chainID, labelNameInAliasOutputMetrics, p.inAliasOutput),
		inOutput:                        createChainMessageMetric(p, chainID, labelNameInOutputMetrics, p.inOutput),
		inOnLedgerRequest:               createChainMessageMetric(p, chainID, labelNameInOnLedgerRequestMetrics, p.inOnLedgerRequest),
		inTxInclusionState:              createChainMessageMetric(p, chainID, labelNameInTxInclusionStateMetrics, p.inTxInclusionState),
		outPublishStateTransaction:      createChainMessageMetric(p, chainID, labelNameOutPublishStateTransactionMetrics, p.outPublishStateTransaction),
		outPublishGovernanceTransaction: createChainMessageMetric(p, chainID, labelNameOutPublishGovernanceTransactionMetrics, p.outPublishGovernanceTransaction),
		outPullLatestOutput:             createChainMessageMetric(p, chainID, labelNameOutPullLatestOutputMetrics, p.outPullLatestOutput),
		outPullTxInclusionState:         createChainMessageMetric(p, chainID, labelNameOutPullTxInclusionStateMetrics, p.outPullTxInclusionState),
		outPullOutputByID:               createChainMessageMetric(p, chainID, labelNameOutPullOutputByIDMetrics, p.outPullOutputByID),
	}
}

func (p *ChainMessageMetricsProvider) InMilestone() IMessageMetric[*nodeclient.MilestoneInfo] {
	return p.inMilestone
}

func (p *ChainMessageMetricsProvider) InStateOutput() IMessageMetric[*InStateOutput] {
	return p.inStateOutput
}

func (p *ChainMessageMetricsProvider) InAliasOutput() IMessageMetric[*iotago.AliasOutput] {
	return p.inAliasOutput
}

func (p *ChainMessageMetricsProvider) InOutput() IMessageMetric[*InOutput] {
	return p.inOutput
}

func (p *ChainMessageMetricsProvider) InOnLedgerRequest() IMessageMetric[isc.OnLedgerRequest] {
	return p.inOnLedgerRequest
}

func (p *ChainMessageMetricsProvider) InTxInclusionState() IMessageMetric[*TxInclusionStateMsg] {
	return p.inTxInclusionState
}

func (p *ChainMessageMetricsProvider) OutPublishStateTransaction() IMessageMetric[*StateTransaction] {
	return p.outPublishStateTransaction
}

func (p *ChainMessageMetricsProvider) OutPublishGovernanceTransaction() IMessageMetric[*iotago.Transaction] {
	return p.outPublishGovernanceTransaction
}

func (p *ChainMessageMetricsProvider) OutPullLatestOutput() IMessageMetric[interface{}] {
	return p.outPullLatestOutput
}

func (p *ChainMessageMetricsProvider) OutPullTxInclusionState() IMessageMetric[iotago.TransactionID] {
	return p.outPullTxInclusionState
}

func (p *ChainMessageMetricsProvider) OutPullOutputByID() IMessageMetric[iotago.OutputID] {
	return p.outPullOutputByID
}

type MessageMetric[T any] struct {
	collectors      *ChainMessageMetricsProvider
	labels          prometheus.Labels
	messagesCount   atomic.Uint32
	lastMessage     T
	lastMessageTime time.Time
}

func newMessageMetric[T any](collectors *ChainMessageMetricsProvider, msgType string) *MessageMetric[T] {
	labels := prometheus.Labels{
		labelNameMessageType: msgType,
	}

	// init values so they appear in prometheus
	collectors.messagesL1.With(labels)
	collectors.lastL1MessageTime.With(labels)

	return &MessageMetric[T]{
		collectors: collectors,
		labels:     labels,
	}
}

func (m *MessageMetric[T]) IncMessages(msg T, ts ...time.Time) {
	timestamp := time.Now()
	if len(ts) > 0 {
		timestamp = ts[0]
	}

	m.messagesCount.Add(1)
	m.lastMessage = msg
	m.lastMessageTime = timestamp

	m.collectors.messagesL1.With(m.labels).Inc()
	m.collectors.lastL1MessageTime.With(m.labels).Set(float64(timestamp.Unix()))
}

func (m *MessageMetric[T]) MessagesTotal() uint32 {
	return m.messagesCount.Load()
}

func (m *MessageMetric[T]) LastMessageTime() time.Time {
	return m.lastMessageTime
}

func (m *MessageMetric[T]) LastMessage() T {
	return m.lastMessage
}

type StateTransaction struct {
	StateIndex  uint32
	Transaction *iotago.Transaction
}

type InStateOutput struct {
	OutputID iotago.OutputID
	Output   iotago.Output
}

type InOutput struct {
	OutputID iotago.OutputID
	Output   iotago.Output
}

type TxInclusionStateMsg struct {
	TxID  iotago.TransactionID
	State string
}

type ChainMessageMetric[T any] struct {
	collectors         *ChainMessageMetricsProvider
	labels             prometheus.Labels
	messageMetricTotal *MessageMetric[T]
	messagesCount      atomic.Uint32
	lastMessage        T
	lastMessageTime    time.Time
}

func createChainMessageMetric[T any](collectors *ChainMessageMetricsProvider, chainID isc.ChainID, msgType string, messageMetricTotal *MessageMetric[T]) *ChainMessageMetric[T] {
	labels := getChainLabels(chainID)
	labels[labelNameMessageType] = msgType

	// init values so they appear in prometheus
	collectors.messagesL1Chain.With(labels)
	collectors.lastL1MessageTimeChain.With(labels)

	return &ChainMessageMetric[T]{
		collectors:         collectors,
		labels:             labels,
		messageMetricTotal: messageMetricTotal,
	}
}

func (m *ChainMessageMetric[T]) IncMessages(msg T, ts ...time.Time) {
	timestamp := time.Now()
	if len(ts) > 0 {
		timestamp = ts[0]
	}
	m.messageMetricTotal.IncMessages(msg, timestamp)

	m.messagesCount.Add(1)
	m.lastMessage = msg
	m.lastMessageTime = timestamp

	m.collectors.messagesL1Chain.With(m.labels).Inc()
	m.collectors.lastL1MessageTimeChain.With(m.labels).Set(float64(timestamp.Unix()))
}

func (m *ChainMessageMetric[T]) MessagesTotal() uint32 {
	return m.messagesCount.Load()
}

func (m *ChainMessageMetric[T]) LastMessageTime() time.Time {
	return m.lastMessageTime
}

func (m *ChainMessageMetric[T]) LastMessage() T {
	return m.lastMessage
}

type ChainMessageMetrics struct {
	inStateOutput      *ChainMessageMetric[*InStateOutput]
	inAliasOutput      *ChainMessageMetric[*iotago.AliasOutput]
	inOutput           *ChainMessageMetric[*InOutput]
	inOnLedgerRequest  *ChainMessageMetric[isc.OnLedgerRequest]
	inTxInclusionState *ChainMessageMetric[*TxInclusionStateMsg]

	outPublishStateTransaction      *ChainMessageMetric[*StateTransaction]
	outPublishGovernanceTransaction *ChainMessageMetric[*iotago.Transaction]
	outPullLatestOutput             *ChainMessageMetric[interface{}]
	outPullTxInclusionState         *ChainMessageMetric[iotago.TransactionID]
	outPullOutputByID               *ChainMessageMetric[iotago.OutputID]
}

func (m *ChainMessageMetrics) InStateOutput() IMessageMetric[*InStateOutput] {
	return m.inStateOutput
}

func (m *ChainMessageMetrics) InAliasOutput() IMessageMetric[*iotago.AliasOutput] {
	return m.inAliasOutput
}

func (m *ChainMessageMetrics) InOutput() IMessageMetric[*InOutput] {
	return m.inOutput
}

func (m *ChainMessageMetrics) InOnLedgerRequest() IMessageMetric[isc.OnLedgerRequest] {
	return m.inOnLedgerRequest
}

func (m *ChainMessageMetrics) InTxInclusionState() IMessageMetric[*TxInclusionStateMsg] {
	return m.inTxInclusionState
}

func (m *ChainMessageMetrics) OutPublishStateTransaction() IMessageMetric[*StateTransaction] {
	return m.outPublishStateTransaction
}

func (m *ChainMessageMetrics) OutPublishGovernanceTransaction() IMessageMetric[*iotago.Transaction] {
	return m.outPublishGovernanceTransaction
}

func (m *ChainMessageMetrics) OutPullLatestOutput() IMessageMetric[interface{}] {
	return m.outPullLatestOutput
}

func (m *ChainMessageMetrics) OutPullTxInclusionState() IMessageMetric[iotago.TransactionID] {
	return m.outPullTxInclusionState
}

func (m *ChainMessageMetrics) OutPullOutputByID() IMessageMetric[iotago.OutputID] {
	return m.outPullOutputByID
}
