package metrics

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/prometheus/client_golang/prometheus"
)

// type StateManagerMetrics interface {
// }
//

type ChainMetrics interface {
	CountMessages()
	CountRequestAckMessages()
	MempoolMetrics
	ConsensusMetrics
}

type MempoolMetrics interface {
	CountOffLedgerRequestIn()
	CountOnLedgerRequestIn()
	CountRequestOut()
	RequestProcessingTime(iscp.RequestID, time.Duration)
}

type ConsensusMetrics interface {
	RecordVMRunTime(time.Duration)
}

type chainMetricsObj struct {
	metrics *Metrics
	chainID *iscp.ChainID
}

var (
	_ ChainMetrics = &chainMetricsObj{}
	_ ChainMetrics = &defaultChainMetrics{}
)

func (c *chainMetricsObj) CountOffLedgerRequestIn() {
	c.metrics.offLedgerRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CountOnLedgerRequestIn() {
	c.metrics.onLedgerRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CountRequestOut() {
	c.metrics.processedRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CountMessages() {
	c.metrics.messagesReceived.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CountRequestAckMessages() {
	c.metrics.requestAckMessages.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) RequestProcessingTime(reqID iscp.RequestID, elapse time.Duration) {
	c.metrics.requestProcessingTime.With(prometheus.Labels{"chain": c.chainID.String(), "request": reqID.String()}).Set(elapse.Seconds())
}

func (c *chainMetricsObj) RecordVMRunTime(elapse time.Duration) {
	c.metrics.vmRunTime.With(prometheus.Labels{"chain": c.chainID.String()}).Set(elapse.Seconds())
}

type defaultChainMetrics struct{}

func DefaultChainMetrics() ChainMetrics {
	return &defaultChainMetrics{}
}

func (m *defaultChainMetrics) CountOffLedgerRequestIn() {}

func (m *defaultChainMetrics) CountOnLedgerRequestIn() {}

func (m *defaultChainMetrics) CountRequestOut() {}

func (m *defaultChainMetrics) CountMessages() {}

func (m *defaultChainMetrics) CountRequestAckMessages() {}

func (m *defaultChainMetrics) RequestProcessingTime(_ iscp.RequestID, _ time.Duration) {}

func (m *defaultChainMetrics) RecordVMRunTime(_ time.Duration) {}
