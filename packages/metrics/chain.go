package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type StateManagerMetrics interface {
	RecordBlockSize(blockIndex uint32, size float64)
	LastSeenStateIndex(stateIndex uint32)
}

type ChainMetrics interface {
	CountMessages()
	CurrentStateIndex(stateIndex uint32)
	MempoolMetrics
	ConsensusMetrics
	StateManagerMetrics
}

type MempoolMetrics interface {
	CountRequestIn(isc.Request)
	CountRequestOut()
	RecordRequestProcessingTime(isc.RequestID, time.Duration)
	CountBlocksPerChain()
}

type ConsensusMetrics interface {
	RecordVMRunTime(time.Duration)
	CountVMRuns()
}

type chainMetricsObj struct {
	metrics *Metrics
	chainID *isc.ChainID
}

var (
	_ ChainMetrics = &chainMetricsObj{}
	_ ChainMetrics = &emptyChainMetrics{}
)

func (c *chainMetricsObj) CountRequestIn(req isc.Request) {
	if req.IsOffLedger() {
		c.metrics.offLedgerRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
	} else {
		c.metrics.onLedgerRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
	}
}

func (c *chainMetricsObj) CountRequestOut() {
	c.metrics.processedRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CountMessages() {
	c.metrics.messagesReceived.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CurrentStateIndex(stateIndex uint32) {
	c.metrics.currentStateIndex.With(prometheus.Labels{"chain": c.chainID.String()}).Set(float64(stateIndex))
}

func (c *chainMetricsObj) RecordRequestProcessingTime(reqID isc.RequestID, elapse time.Duration) {
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

type emptyChainMetrics struct{}

func EmptyChainMetrics() ChainMetrics {
	return &emptyChainMetrics{}
}

func (m *emptyChainMetrics) CountRequestIn(_ isc.Request) {}

func (m *emptyChainMetrics) CountRequestOut() {}

func (m *emptyChainMetrics) CountMessages() {}

func (m *emptyChainMetrics) CurrentStateIndex(stateIndex uint32) {}

func (m *emptyChainMetrics) RecordRequestProcessingTime(_ isc.RequestID, _ time.Duration) {}

func (m *emptyChainMetrics) RecordVMRunTime(_ time.Duration) {}

func (m *emptyChainMetrics) CountVMRuns() {}

func (m *emptyChainMetrics) CountBlocksPerChain() {}

func (m *emptyChainMetrics) RecordBlockSize(_ uint32, _ float64) {}

func (m *emptyChainMetrics) LastSeenStateIndex(stateIndex uint32) {}
