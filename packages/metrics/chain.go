package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type StateManagerMetrics interface {
	SetBlockSize(blockIndex uint32, size float64)
	SetLastSeenStateIndex(stateIndex uint32)
}

type ChainMetrics interface {
	IncRequestsAckMessages()
	IncMessagesReceived()
	SetCurrentStateIndex(stateIndex uint32)
	MempoolMetrics
	ConsensusMetrics
	StateManagerMetrics
}

type MempoolMetrics interface {
	IncRequestsReceived(isc.Request)
	IncRequestsProcessed()
	SetRequestProcessingTime(isc.RequestID, time.Duration)
	IncBlocksPerChain()
}

type ConsensusMetrics interface {
	SetVMRunTime(time.Duration)
	IncVMRunsCounter()
}

type chainMetricsObj struct {
	metrics *Metrics
	chainID isc.ChainID
}

var (
	_ ChainMetrics = &chainMetricsObj{}
	_ ChainMetrics = &emptyChainMetrics{}
)

func (c *chainMetricsObj) IncRequestsReceived(req isc.Request) {
	if req.IsOffLedger() {
		c.metrics.requestsReceivedOffLedger.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
	} else {
		c.metrics.requestsReceivedOnLedger.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
	}
}

func (c *chainMetricsObj) IncRequestsProcessed() {
	c.metrics.requestsProcessed.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) IncRequestsAckMessages() {
	c.metrics.requestsAckMessages.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) SetRequestProcessingTime(reqID isc.RequestID, elapse time.Duration) {
	c.metrics.requestsProcessingTime.With(prometheus.Labels{"chain": c.chainID.String(), "request": reqID.String()}).Set(elapse.Seconds())
}

func (c *chainMetricsObj) IncMessagesReceived() {
	c.metrics.messagesReceived.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) SetVMRunTime(elapse time.Duration) {
	c.metrics.vmRunTime.With(prometheus.Labels{"chain": c.chainID.String()}).Set(elapse.Seconds())
}

func (c *chainMetricsObj) IncVMRunsCounter() {
	c.metrics.vmRunsTotal.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) IncBlocksPerChain() {
	c.metrics.blocksTotalPerChain.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) SetBlockSize(blockIndex uint32, blockSize float64) {
	c.metrics.blockSizesPerChain.With(prometheus.Labels{"chain": c.chainID.String(), "block_index": fmt.Sprintf("%d", blockIndex)}).Set(blockSize)
}

func (c *chainMetricsObj) SetCurrentStateIndex(stateIndex uint32) {
	c.metrics.stateIndexCurrent.With(prometheus.Labels{"chain": c.chainID.String()}).Set(float64(stateIndex))
}

func (c *chainMetricsObj) SetLastSeenStateIndex(stateIndex uint32) {
	if c.metrics.lastSeenStateIndexVal >= stateIndex {
		return
	}
	c.metrics.lastSeenStateIndexVal = stateIndex
	c.metrics.stateIndexLatestSeen.With(prometheus.Labels{"chain": c.chainID.String()}).Set(float64(stateIndex))
}

type emptyChainMetrics struct{}

func EmptyChainMetrics() ChainMetrics {
	return &emptyChainMetrics{}
}

func (m *emptyChainMetrics) IncRequestsReceived(_ isc.Request) {}

func (m *emptyChainMetrics) IncRequestsProcessed() {}

func (m *emptyChainMetrics) IncRequestsAckMessages() {}

func (m *emptyChainMetrics) SetRequestProcessingTime(_ isc.RequestID, _ time.Duration) {}

func (m *emptyChainMetrics) IncMessagesReceived() {}

func (m *emptyChainMetrics) SetVMRunTime(_ time.Duration) {}

func (m *emptyChainMetrics) IncVMRunsCounter() {}

func (m *emptyChainMetrics) IncBlocksPerChain() {}

func (m *emptyChainMetrics) SetBlockSize(_ uint32, _ float64) {}

func (m *emptyChainMetrics) SetCurrentStateIndex(stateIndex uint32) {}

func (m *emptyChainMetrics) SetLastSeenStateIndex(stateIndex uint32) {}
