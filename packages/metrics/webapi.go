package metrics

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
)

type IWebAPIMetrics interface {
	WebAPIRequest(operation string, httpStatusCode int, duration time.Duration)
	EvmRPCCall(operation string, success bool, duration time.Duration)
}

var (
	_ IWebAPIMetrics = &emptyWebAPIChainMetrics{}
	_ IWebAPIMetrics = &webAPIChainMetrics{}
)

type emptyWebAPIChainMetrics struct{}

func NewEmptyWebAPIMetrics() IWebAPIMetrics {
	return &emptyWebAPIChainMetrics{}
}

func (m *emptyWebAPIChainMetrics) WebAPIRequest(operation string, httpStatusCode int, duration time.Duration) {
}

func (m *emptyWebAPIChainMetrics) EvmRPCCall(operation string, success bool, duration time.Duration) {
}

type webAPIChainMetrics struct {
	provider *ChainMetricsProvider
	chainID  isc.ChainID
}

func newWebAPIChainMetrics(provider *ChainMetricsProvider, chainID isc.ChainID) *webAPIChainMetrics {
	return &webAPIChainMetrics{
		provider: provider,
		chainID:  chainID,
	}
}

func (m *webAPIChainMetrics) WebAPIRequest(operation string, httpStatusCode int, duration time.Duration) {
	labels := getChainLabels(m.chainID)
	labels[labelNameWebapiRequestOperation] = operation
	labels[labelNameWebapiRequestStatusCode] = fmt.Sprintf("%d", httpStatusCode)
	m.provider.webAPIRequests.With(labels).Observe(duration.Seconds())
}

func (m *webAPIChainMetrics) EvmRPCCall(operation string, success bool, duration time.Duration) {
	labels := getChainLabels(m.chainID)
	labels[labelNameWebapiRequestOperation] = operation
	labels[labelNameWebapiEvmRPCSuccess] = fmt.Sprintf("%v", success)
	m.provider.webAPIEvmRPCCalls.With(labels).Observe(duration.Seconds())
}
