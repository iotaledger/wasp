package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainMempoolMetrics interface {
	IncBlocksPerChain()
	IncRequestsReceived(isc.Request)
	IncRequestsProcessed()
	IncRequestsAckMessages()
	SetRequestProcessingTime(time.Duration)
}

var (
	_ IChainMempoolMetrics = &emptyChainMempoolMetric{}
	_ IChainMempoolMetrics = &chainMempoolMetric{}
)

type emptyChainMempoolMetric struct{}

func NewEmptyChainMempoolMetric() IChainMempoolMetrics                    { return &emptyChainMempoolMetric{} }
func (m *emptyChainMempoolMetric) IncBlocksPerChain()                     {}
func (m *emptyChainMempoolMetric) IncRequestsReceived(isc.Request)        {}
func (m *emptyChainMempoolMetric) IncRequestsProcessed()                  {}
func (m *emptyChainMempoolMetric) IncRequestsAckMessages()                {}
func (m *emptyChainMempoolMetric) SetRequestProcessingTime(time.Duration) {}

type chainMempoolMetric struct {
	provider      *ChainMetricsProvider
	metricsLabels prometheus.Labels
}

func newChainMempoolMetric(provider *ChainMetricsProvider, chainID isc.ChainID) *chainMempoolMetric {
	metricsLabels := getChainLabels(chainID)

	// init values so they appear in prometheus
	provider.blocksTotalPerChain.With(metricsLabels)
	provider.requestsReceivedOffLedger.With(metricsLabels)
	provider.requestsReceivedOnLedger.With(metricsLabels)
	provider.requestsProcessed.With(metricsLabels)
	provider.requestsAckMessages.With(metricsLabels)
	provider.requestsProcessingTime.With(metricsLabels)

	return &chainMempoolMetric{
		provider:      provider,
		metricsLabels: metricsLabels,
	}
}

func (m *chainMempoolMetric) IncBlocksPerChain() {
	m.provider.blocksTotalPerChain.With(m.metricsLabels).Inc()
}

func (m *chainMempoolMetric) IncRequestsReceived(request isc.Request) {
	if request.IsOffLedger() {
		m.provider.requestsReceivedOffLedger.With(m.metricsLabels).Inc()
	} else {
		m.provider.requestsReceivedOnLedger.With(m.metricsLabels).Inc()
	}
}

func (m *chainMempoolMetric) IncRequestsProcessed() {
	m.provider.requestsProcessed.With(m.metricsLabels).Inc()
}

func (m *chainMempoolMetric) IncRequestsAckMessages() {
	m.provider.requestsAckMessages.With(m.metricsLabels).Inc()
}

func (m *chainMempoolMetric) SetRequestProcessingTime(duration time.Duration) {
	m.provider.requestsProcessingTime.With(m.metricsLabels).Set(duration.Seconds())
}
