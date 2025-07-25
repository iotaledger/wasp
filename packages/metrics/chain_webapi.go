package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainWebAPIMetricsProvider struct {
	requests    *prometheus.HistogramVec
	evmRPCCalls *prometheus.HistogramVec
}

func NewChainWebAPIMetricsProvider() *ChainWebAPIMetricsProvider {
	return &ChainWebAPIMetricsProvider{
		requests: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "webapi",
			Name:      "webapi_requests",
			Help:      "Time elapsed (s) processing requests",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain, labelNameWebapiRequestOperation, labelNameWebapiRequestStatusCode}),
		evmRPCCalls: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "iota_wasp",
			Subsystem: "webapi",
			Name:      "webapi_evm_rpc_calls",
			Help:      "Time elapsed (s) processing evm rpc requests",
			Buckets:   execTimeBuckets,
		}, []string{labelNameChain, labelNameWebapiRequestOperation, labelNameWebapiEvmRPCSuccess}),
	}
}

func (p *ChainWebAPIMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.requests,
		p.evmRPCCalls,
	)
}

func (p *ChainWebAPIMetricsProvider) CreateForChain(chainID isc.ChainID) *ChainWebAPIMetrics {
	return newChainWebAPIMetrics(p, chainID)
}

type ChainWebAPIMetrics struct {
	collectors *ChainWebAPIMetricsProvider
	chainID    isc.ChainID
}

func newChainWebAPIMetrics(collectors *ChainWebAPIMetricsProvider, chainID isc.ChainID) *ChainWebAPIMetrics {
	return &ChainWebAPIMetrics{
		collectors: collectors,
		chainID:    chainID,
	}
}

func (m *ChainWebAPIMetrics) WebAPIRequest(operation string, httpStatusCode int, duration time.Duration) {
	labels := getChainLabels(m.chainID)
	labels[labelNameWebapiRequestOperation] = operation
	labels[labelNameWebapiRequestStatusCode] = fmt.Sprintf("%d", httpStatusCode)
	m.collectors.requests.With(labels).Observe(duration.Seconds())
}

func (m *ChainWebAPIMetrics) EVMRPCCall(operation string, success bool, duration time.Duration) {
	labels := getChainLabels(m.chainID)
	labels[labelNameWebapiRequestOperation] = operation
	labels[labelNameWebapiEvmRPCSuccess] = fmt.Sprintf("%v", success)
	m.collectors.evmRPCCalls.With(labels).Observe(duration.Seconds())
}
