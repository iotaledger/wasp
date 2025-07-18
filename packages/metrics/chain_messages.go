package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type IMessageMetric[T any] interface {
	IncMessages(msg T, ts ...time.Time)
	MessagesTotal() uint32
	LastMessageTime() time.Time
	LastMessage() T
}

type ChainMessageMetricsProvider struct {
	messagesL1             *prometheus.CounterVec
	lastL1MessageTime      *prometheus.GaugeVec
	messagesL1Chain        *prometheus.CounterVec
	lastL1MessageTimeChain *prometheus.GaugeVec

	inAnchor                   *MessageMetric[*StateAnchor]
	inOnLedgerRequest          *MessageMetric[isc.OnLedgerRequest]
	outPublishStateTransaction *MessageMetric[*StateTransaction]
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

	p.inAnchor = newMessageMetric[*StateAnchor](p, labelNameInAnchorMetrics)
	p.inOnLedgerRequest = newMessageMetric[isc.OnLedgerRequest](p, labelNameInOnLedgerRequestMetrics)
	p.outPublishStateTransaction = newMessageMetric[*StateTransaction](p, labelNameOutPublishStateTransactionMetrics)

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
		inAnchor:                   createChainMessageMetric(p, chainID, labelNameInAnchorMetrics, p.inAnchor),
		inOnLedgerRequest:          createChainMessageMetric(p, chainID, labelNameInOnLedgerRequestMetrics, p.inOnLedgerRequest),
		outPublishStateTransaction: createChainMessageMetric(p, chainID, labelNameOutPublishStateTransactionMetrics, p.outPublishStateTransaction),
	}
}

func (p *ChainMessageMetricsProvider) InAnchor() IMessageMetric[*StateAnchor] {
	return p.inAnchor
}

func (p *ChainMessageMetricsProvider) InOnLedgerRequest() IMessageMetric[isc.OnLedgerRequest] {
	return p.inOnLedgerRequest
}

func (p *ChainMessageMetricsProvider) OutPublishStateTransaction() IMessageMetric[*StateTransaction] {
	return p.outPublishStateTransaction
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

type StateAnchor struct {
	StateIndex    uint32
	StateMetadata string
	Ref           iotago.ObjectRef
}

type StateTransaction struct {
	StateIndex        uint32
	TransactionDigest iotago.TransactionDigest
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
	inAnchor                   *ChainMessageMetric[*StateAnchor]
	inOnLedgerRequest          *ChainMessageMetric[isc.OnLedgerRequest]
	outPublishStateTransaction *ChainMessageMetric[*StateTransaction]
}

func (m *ChainMessageMetrics) InAnchor() IMessageMetric[*StateAnchor] {
	return m.inAnchor
}

func (m *ChainMessageMetrics) InOnLedgerRequest() IMessageMetric[isc.OnLedgerRequest] {
	return m.inOnLedgerRequest
}

func (m *ChainMessageMetrics) OutPublishStateTransaction() IMessageMetric[*StateTransaction] {
	return m.outPublishStateTransaction
}
