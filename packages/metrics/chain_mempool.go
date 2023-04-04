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

	SetTimePoolSize(count int)
	SetOnLedgerPoolSize(count int)
	SetOffLedgerPoolSize(count int)
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
func (m *emptyChainMempoolMetric) SetTimePoolSize(count int)              {}
func (m *emptyChainMempoolMetric) SetOnLedgerPoolSize(count int)          {}
func (m *emptyChainMempoolMetric) SetOffLedgerPoolSize(count int)         {}

type chainMempoolMetric struct {
	provider      *ChainMetricsProvider
	metricsLabels prometheus.Labels

	vTimePoolSize      int
	vOnLedgerPoolSize  int
	vOffLedgerPoolSize int
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

	provider.mempoolTimePoolSize.With(metricsLabels)
	provider.mempoolOnLedgerPoolSize.With(metricsLabels)
	provider.mempoolOffLedgerPoolSize.With(metricsLabels)
	provider.mempoolTotalSize.With(metricsLabels)

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

func (m *chainMempoolMetric) SetTimePoolSize(count int) {
	m.vTimePoolSize = count
	m.provider.mempoolTimePoolSize.With(m.metricsLabels).Set(float64(count))
	m.deriveTotalPoolSize()
}

func (m *chainMempoolMetric) SetOnLedgerPoolSize(count int) {
	m.vOnLedgerPoolSize = count
	m.provider.mempoolOnLedgerPoolSize.With(m.metricsLabels).Set(float64(count))
	m.deriveTotalPoolSize()
}

func (m *chainMempoolMetric) SetOffLedgerPoolSize(count int) {
	m.vOffLedgerPoolSize = count
	m.provider.mempoolOffLedgerPoolSize.With(m.metricsLabels).Set(float64(count))
	m.deriveTotalPoolSize()
}

func (m *chainMempoolMetric) deriveTotalPoolSize() {
	m.provider.mempoolTotalSize.With(m.metricsLabels).Set(float64(m.vTimePoolSize + m.vOnLedgerPoolSize + m.vOffLedgerPoolSize))
}
