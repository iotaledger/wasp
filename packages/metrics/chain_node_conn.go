package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainNodeConnMetrics interface {
	L1RequestReceived()
	L1AliasOutputReceived()
	TXPublishStarted()
	TXPublishResult(confirmed bool, duration time.Duration)
}

var (
	_ IChainNodeConnMetrics = &emptyChainNodeConnMetric{}
	_ IChainNodeConnMetrics = &chainNodeConnMetric{}
)

type emptyChainNodeConnMetric struct{}

func NewEmptyChainNodeConnMetric() IChainNodeConnMetrics {
	return &emptyChainNodeConnMetric{}
}
func (m *emptyChainNodeConnMetric) L1RequestReceived()                                     {}
func (m *emptyChainNodeConnMetric) L1AliasOutputReceived()                                 {}
func (m *emptyChainNodeConnMetric) TXPublishStarted()                                      {}
func (m *emptyChainNodeConnMetric) TXPublishResult(confirmed bool, duration time.Duration) {}

type chainNodeConnMetric struct {
	ncL1RequestReceived     prometheus.Counter
	ncL1AliasOutputReceived prometheus.Counter
	ncTXPublishStarted      prometheus.Counter
	ncTXPublishResult       map[bool]prometheus.Observer
}

func newChainNodeConnMetric(provider *ChainMetricsProvider, chainID isc.ChainID) *chainNodeConnMetric {
	chainLabels := getChainLabels(chainID)
	return &chainNodeConnMetric{
		ncL1RequestReceived:     provider.ncL1RequestReceived.With(chainLabels),
		ncL1AliasOutputReceived: provider.ncL1AliasOutputReceived.With(chainLabels),
		ncTXPublishStarted:      provider.ncTXPublishStarted.With(chainLabels),
		ncTXPublishResult: map[bool]prometheus.Observer{
			true:  provider.ncTXPublishResult.MustCurryWith(chainLabels).With(prometheus.Labels{labelTxPublishResult: "confirmed"}),
			false: provider.ncTXPublishResult.MustCurryWith(chainLabels).With(prometheus.Labels{labelTxPublishResult: "rejected"}),
		},
	}
}

func (m *chainNodeConnMetric) L1RequestReceived() {
	m.ncL1RequestReceived.Inc()
}

func (m *chainNodeConnMetric) L1AliasOutputReceived() {
	m.ncL1AliasOutputReceived.Inc()
}

func (m *chainNodeConnMetric) TXPublishStarted() {
	m.ncTXPublishStarted.Inc()
}

func (m *chainNodeConnMetric) TXPublishResult(confirmed bool, duration time.Duration) {
	m.ncTXPublishResult[confirmed].Observe(duration.Seconds())
}
