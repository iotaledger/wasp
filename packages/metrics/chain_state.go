package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type IChainStateMetrics interface {
	SetChainActiveStateWant(stateIndex uint32)
	SetChainActiveStateHave(stateIndex uint32)
	SetChainConfirmedStateWant(stateIndex uint32)
	SetChainConfirmedStateHave(stateIndex uint32)
}

var (
	_ IChainStateMetrics = &emptyChainStateMetric{}
	_ IChainStateMetrics = &chainStateMetric{}
)

type emptyChainStateMetric struct{}

func NewEmptyChainStateMetric() IChainStateMetrics {
	return &emptyChainStateMetric{}
}

func (m *emptyChainStateMetric) SetChainActiveStateWant(stateIndex uint32)    {}
func (m *emptyChainStateMetric) SetChainActiveStateHave(stateIndex uint32)    {}
func (m *emptyChainStateMetric) SetChainConfirmedStateWant(stateIndex uint32) {}
func (m *emptyChainStateMetric) SetChainConfirmedStateHave(stateIndex uint32) {}

type chainStateMetric struct {
	chainID               isc.ChainID
	provider              *ChainMetricsProvider
	metricsLabels         prometheus.Labels
	lastSeenStateIndexVal uint32
}

func newChainStateMetric(provider *ChainMetricsProvider, chainID isc.ChainID) *chainStateMetric {
	metricsLabels := getChainLabels(chainID)

	// init values so they appear in prometheus
	provider.chainActiveStateWant.With(metricsLabels)
	provider.chainActiveStateHave.With(metricsLabels)
	provider.chainConfirmedStateWant.With(metricsLabels)
	provider.chainConfirmedStateHave.With(metricsLabels)

	return &chainStateMetric{
		chainID:               chainID,
		provider:              provider,
		metricsLabels:         metricsLabels,
		lastSeenStateIndexVal: 0,
	}
}

func (m *chainStateMetric) SetChainActiveStateWant(stateIndex uint32) {
	m.provider.chainActiveStateWant.With(m.metricsLabels).Set(float64(stateIndex))
}

func (m *chainStateMetric) SetChainActiveStateHave(stateIndex uint32) {
	m.provider.chainActiveStateHave.With(m.metricsLabels).Set(float64(stateIndex))
}

func (m *chainStateMetric) SetChainConfirmedStateWant(stateIndex uint32) {
	m.provider.chainConfirmedStateWant.With(m.metricsLabels).Set(float64(stateIndex))
	m.provider.chainConfirmedStateLag.Want(m.chainID, stateIndex)
}

func (m *chainStateMetric) SetChainConfirmedStateHave(stateIndex uint32) {
	m.provider.chainConfirmedStateHave.With(m.metricsLabels).Set(float64(stateIndex))
	m.provider.chainConfirmedStateLag.Have(m.chainID, stateIndex)
}
