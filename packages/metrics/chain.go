package metrics

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/prometheus/client_golang/prometheus"
)

// type StateManagerMetrics interface {
// }
//
// type ConsensusMetrics interface {
// }

type ChainMetrics interface {
	MempoolMetrics
}

type MempoolMetrics interface {
	CountOffLedgerRequestIn()
	CountOnLedgerRequestIn()
	CountRequestOut()
}

type chainMetricsObj struct {
	metrics *Metrics
	chainID *iscp.ChainID
}

func (c *chainMetricsObj) CountOffLedgerRequestIn() {
	c.metrics.offLedgerRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CountOnLedgerRequestIn() {
	c.metrics.onLedgerRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

func (c *chainMetricsObj) CountRequestOut() {
	c.metrics.processedRequestCounter.With(prometheus.Labels{"chain": c.chainID.String()}).Inc()
}

type defaultChainMetrics struct{}

func DefaultChainMetrics() ChainMetrics {
	return &defaultChainMetrics{}
}

func (m *defaultChainMetrics) CountOffLedgerRequestIn() {}

func (m *defaultChainMetrics) CountOnLedgerRequestIn() {}

func (m *defaultChainMetrics) CountRequestOut() {}
