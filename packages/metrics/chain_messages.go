package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type IMessageMetric[T any] interface {
	IncMessages(msg T, ts ...time.Time)
	MessagesTotal() uint32
	LastMessageTime() time.Time
	LastMessage() T
}

type IChainMessageMetrics interface {
	InStateOutput() IMessageMetric[*InStateOutput]
	InAliasOutput() IMessageMetric[*iotago.AliasOutput]
	InOutput() IMessageMetric[*InOutput]
	InOnLedgerRequest() IMessageMetric[isc.OnLedgerRequest]
	InTxInclusionState() IMessageMetric[*TxInclusionStateMsg]
	OutPublishStateTransaction() IMessageMetric[*StateTransaction]
	OutPublishGovernanceTransaction() IMessageMetric[*iotago.Transaction]
	OutPullLatestOutput() IMessageMetric[interface{}]
	OutPullTxInclusionState() IMessageMetric[iotago.TransactionID]
	OutPullOutputByID() IMessageMetric[iotago.OutputID]
}

var (
	_ IMessageMetric[any]  = &messageMetric[any]{}
	_ IMessageMetric[any]  = &emptyChainMessageMetric[any]{}
	_ IMessageMetric[any]  = &chainMessageMetric[any]{}
	_ IChainMessageMetrics = &emptyChainMessageMetrics{}
	_ IChainMessageMetrics = &chainMessageMetrics{}
)

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

type emptyChainMessageMetric[T any] struct{}

func (m emptyChainMessageMetric[T]) IncMessages(msg T, ts ...time.Time) {}
func (m emptyChainMessageMetric[T]) MessagesTotal() uint32              { return 0 }
func (m emptyChainMessageMetric[T]) LastMessageTime() time.Time         { return time.Time{} }
func (m emptyChainMessageMetric[T]) LastMessage() T                     { return *new(T) }

type emptyChainMessageMetrics struct{}

func NewEmptyChainMessageMetrics() IChainMessageMetrics { return &emptyChainMessageMetrics{} }
func (m *emptyChainMessageMetrics) InStateOutput() IMessageMetric[*InStateOutput] {
	return &emptyChainMessageMetric[*InStateOutput]{}
}

func (m *emptyChainMessageMetrics) InAliasOutput() IMessageMetric[*iotago.AliasOutput] {
	return &emptyChainMessageMetric[*iotago.AliasOutput]{}
}

func (m *emptyChainMessageMetrics) InOutput() IMessageMetric[*InOutput] {
	return &emptyChainMessageMetric[*InOutput]{}
}

func (m *emptyChainMessageMetrics) InOnLedgerRequest() IMessageMetric[isc.OnLedgerRequest] {
	return &emptyChainMessageMetric[isc.OnLedgerRequest]{}
}

func (m *emptyChainMessageMetrics) InTxInclusionState() IMessageMetric[*TxInclusionStateMsg] {
	return &emptyChainMessageMetric[*TxInclusionStateMsg]{}
}

func (m *emptyChainMessageMetrics) OutPublishStateTransaction() IMessageMetric[*StateTransaction] {
	return &emptyChainMessageMetric[*StateTransaction]{}
}

func (m *emptyChainMessageMetrics) OutPublishGovernanceTransaction() IMessageMetric[*iotago.Transaction] {
	return &emptyChainMessageMetric[*iotago.Transaction]{}
}

func (m *emptyChainMessageMetrics) OutPullLatestOutput() IMessageMetric[interface{}] {
	return &emptyChainMessageMetric[interface{}]{}
}

func (m *emptyChainMessageMetrics) OutPullTxInclusionState() IMessageMetric[iotago.TransactionID] {
	return &emptyChainMessageMetric[iotago.TransactionID]{}
}

func (m *emptyChainMessageMetrics) OutPullOutputByID() IMessageMetric[iotago.OutputID] {
	return &emptyChainMessageMetric[iotago.OutputID]{}
}

type chainMessageMetric[T any] struct {
	provider           *ChainMetricsProvider
	metricsLabels      prometheus.Labels
	messageMetricTotal *messageMetric[T]
	messagesCount      atomic.Uint32
	lastMessage        T
	lastMessageTime    time.Time
}

func createChainMessageMetric[T any](provider *ChainMetricsProvider, chainID isc.ChainID, msgType string, messageMetricTotal *messageMetric[T]) IMessageMetric[T] {
	metricsLabels := getChainMessageTypeLabels(chainID, msgType)

	// init values so they appear in prometheus
	provider.messagesL1Chain.With(metricsLabels)
	provider.lastL1MessageTimeChain.With(metricsLabels)

	return &chainMessageMetric[T]{
		provider:           provider,
		metricsLabels:      metricsLabels,
		messageMetricTotal: messageMetricTotal,
	}
}

func (m *chainMessageMetric[T]) IncMessages(msg T, ts ...time.Time) {
	timestamp := time.Now()
	if len(ts) > 0 {
		timestamp = ts[0]
	}
	m.messageMetricTotal.IncMessages(msg, timestamp)

	m.messagesCount.Add(1)
	m.lastMessage = msg
	m.lastMessageTime = timestamp

	m.provider.messagesL1Chain.With(m.metricsLabels).Inc()
	m.provider.lastL1MessageTimeChain.With(m.metricsLabels).Set(float64(timestamp.Unix()))
}

func (m *chainMessageMetric[T]) MessagesTotal() uint32 {
	return m.messagesCount.Load()
}

func (m *chainMessageMetric[T]) LastMessageTime() time.Time {
	return m.lastMessageTime
}

func (m *chainMessageMetric[T]) LastMessage() T {
	return m.lastMessage
}

type chainMessageMetrics struct {
	inStateOutputMetrics      IMessageMetric[*InStateOutput]
	inAliasOutputMetrics      IMessageMetric[*iotago.AliasOutput]
	inOutputMetrics           IMessageMetric[*InOutput]
	inOnLedgerRequestMetrics  IMessageMetric[isc.OnLedgerRequest]
	inTxInclusionStateMetrics IMessageMetric[*TxInclusionStateMsg]

	outPublishStateTransactionMetrics      IMessageMetric[*StateTransaction]
	outPublishGovernanceTransactionMetrics IMessageMetric[*iotago.Transaction]
	outPullLatestOutputMetrics             IMessageMetric[interface{}]
	outPullTxInclusionStateMetrics         IMessageMetric[iotago.TransactionID]
	outPullOutputByIDMetrics               IMessageMetric[iotago.OutputID]
}

func newChainMessageMetrics(provider *ChainMetricsProvider, chainID isc.ChainID) *chainMessageMetrics {
	return &chainMessageMetrics{
		inStateOutputMetrics:                   createChainMessageMetric(provider, chainID, labelNameInStateOutputMetrics, provider.inStateOutputMetrics),
		inAliasOutputMetrics:                   createChainMessageMetric(provider, chainID, labelNameInAliasOutputMetrics, provider.inAliasOutputMetrics),
		inOutputMetrics:                        createChainMessageMetric(provider, chainID, labelNameInOutputMetrics, provider.inOutputMetrics),
		inOnLedgerRequestMetrics:               createChainMessageMetric(provider, chainID, labelNameInOnLedgerRequestMetrics, provider.inOnLedgerRequestMetrics),
		inTxInclusionStateMetrics:              createChainMessageMetric(provider, chainID, labelNameInTxInclusionStateMetrics, provider.inTxInclusionStateMetrics),
		outPublishStateTransactionMetrics:      createChainMessageMetric(provider, chainID, labelNameOutPublishStateTransactionMetrics, provider.outPublishStateTransactionMetrics),
		outPublishGovernanceTransactionMetrics: createChainMessageMetric(provider, chainID, labelNameOutPublishGovernanceTransactionMetrics, provider.outPublishGovernanceTransactionMetrics),
		outPullLatestOutputMetrics:             createChainMessageMetric(provider, chainID, labelNameOutPullLatestOutputMetrics, provider.outPullLatestOutputMetrics),
		outPullTxInclusionStateMetrics:         createChainMessageMetric(provider, chainID, labelNameOutPullTxInclusionStateMetrics, provider.outPullTxInclusionStateMetrics),
		outPullOutputByIDMetrics:               createChainMessageMetric(provider, chainID, labelNameOutPullOutputByIDMetrics, provider.outPullOutputByIDMetrics),
	}
}

func (m *chainMessageMetrics) InStateOutput() IMessageMetric[*InStateOutput] {
	return m.inStateOutputMetrics
}

func (m *chainMessageMetrics) InAliasOutput() IMessageMetric[*iotago.AliasOutput] {
	return m.inAliasOutputMetrics
}

func (m *chainMessageMetrics) InOutput() IMessageMetric[*InOutput] {
	return m.inOutputMetrics
}

func (m *chainMessageMetrics) InOnLedgerRequest() IMessageMetric[isc.OnLedgerRequest] {
	return m.inOnLedgerRequestMetrics
}

func (m *chainMessageMetrics) InTxInclusionState() IMessageMetric[*TxInclusionStateMsg] {
	return m.inTxInclusionStateMetrics
}

func (m *chainMessageMetrics) OutPublishStateTransaction() IMessageMetric[*StateTransaction] {
	return m.outPublishStateTransactionMetrics
}

func (m *chainMessageMetrics) OutPublishGovernanceTransaction() IMessageMetric[*iotago.Transaction] {
	return m.outPublishGovernanceTransactionMetrics
}

func (m *chainMessageMetrics) OutPullLatestOutput() IMessageMetric[interface{}] {
	return m.outPullLatestOutputMetrics
}

func (m *chainMessageMetrics) OutPullTxInclusionState() IMessageMetric[iotago.TransactionID] {
	return m.outPullTxInclusionStateMetrics
}

func (m *chainMessageMetrics) OutPullOutputByID() IMessageMetric[iotago.OutputID] {
	return m.outPullOutputByIDMetrics
}
