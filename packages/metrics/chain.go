package metrics

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/prometheus/client_golang/prometheus"
)

type StateManagerMetrics interface {
	RecordBlockSize(blockIndex uint32, size float64)
	LastSeenStateIndex(stateIndex uint32)
}

type ChainMetrics interface {
	CountMessages()
	CountRequestAckMessages()
	CurrentStateIndex(stateIndex uint32)
	MempoolMetrics
	ConsensusMetrics
	StateManagerMetrics
}

type MempoolMetrics interface {
	CountOffLedgerRequestIn()
	CountOnLedgerRequestIn()
	CountRequestOut()
	RecordRequestProcessingTime(iscp.RequestID, time.Duration)
	CountBlocksPerChain()
}

type ConsensusMetrics interface {
	RecordVMRunTime(time.Duration)
	CountVMRuns()
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

func (c *chainMetricsObj) CurrentStateIndex(stateIndex uint32) {
	c.metrics.currentStateIndex.With(prometheus.Labels{"chain": c.chainID.String()}).Set(float64(stateIndex))
}

func (c *chainMetricsObj) RecordRequestProcessingTime(reqID iscp.RequestID, elapse time.Duration) {
	c.metrics.requestProcessingTime.With(prometheus.Labels{"chain": c.chainID.String(), "request": reqID.String()}).Set(elapse.Seconds())
}

func (c *chainMetricsObj) RecordVMRunTime(elapse time.Duration) {
	c.metrics.vmRunTime.With(prometheus.Labels{"chain": c.chainID.String()}).Set(elapse.Seconds())
}

func (c *chainMetricsObj) CountVMRuns() {
	c.metrics.vmRunCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CountBlocksPerChain() {
	c.metrics.blocksPerChain.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) RecordBlockSize(blockIndex uint32, blockSize float64) {
	c.metrics.blockSizes.With(prometheus.Labels{"chain": c.chainID.String(), "block_index": fmt.Sprintf("%d", blockIndex)}).Set(blockSize)
}

func (c *chainMetricsObj) LastSeenStateIndex(stateIndex uint32) {
	if c.metrics.lastSeenStateIndexVal >= stateIndex {
		return
	}
	c.metrics.lastSeenStateIndexVal = stateIndex
	c.metrics.lastSeenStateIndex.With(prometheus.Labels{"chain": c.chainID.String()}).Set(float64(stateIndex))
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

func (m *defaultChainMetrics) CurrentStateIndex(stateIndex uint32) {}

func (m *defaultChainMetrics) RecordRequestProcessingTime(_ iscp.RequestID, _ time.Duration) {}

func (m *defaultChainMetrics) RecordVMRunTime(_ time.Duration) {}

func (m *defaultChainMetrics) CountVMRuns() {}

func (m *defaultChainMetrics) CountBlocksPerChain() {}

func (m *defaultChainMetrics) RecordBlockSize(_ uint32, _ float64) {}

func (m *defaultChainMetrics) LastSeenStateIndex(stateIndex uint32) {}
