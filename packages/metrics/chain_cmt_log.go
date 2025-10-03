package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type ChainCommitteeLogMetricsProvider struct {
	logIndexIncReasonConsOut     *prometheus.CounterVec
	logIndexIncReasonRecover     *prometheus.CounterVec
	logIndexIncReasonL1RepAnchor *prometheus.CounterVec
	logIndexIncReasonStarted     *prometheus.CounterVec
}

func newChainCommitteeLogMetricsProvider() *ChainCommitteeLogMetricsProvider {
	return &ChainCommitteeLogMetricsProvider{
		logIndexIncReasonConsOut: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "cmtlog",
			Name:      "log_index_inc_reason_ConsOut",
			Help:      "Number if times LogIndex was increased because of consensus output.",
		}, []string{labelNameChain}),
		logIndexIncReasonRecover: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "cmtlog",
			Name:      "log_index_inc_reason_Recover",
			Help:      "Number if times LogIndex was increased because of recovery procedure.",
		}, []string{labelNameChain}),
		logIndexIncReasonL1RepAnchor: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "cmtlog",
			Name:      "log_index_inc_reason_L1RepAnchor",
			Help:      "Number if times LogIndex was increased because L1 replaced TIP anchor.",
		}, []string{labelNameChain}),
		logIndexIncReasonStarted: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "cmtlog",
			Name:      "log_index_inc_reason_Started",
			Help:      "Number if times LogIndex was increased because other nodes started the consensus.",
		}, []string{labelNameChain}),
	}
}

func (p *ChainCommitteeLogMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.logIndexIncReasonConsOut,
		p.logIndexIncReasonRecover,
		p.logIndexIncReasonL1RepAnchor,
		p.logIndexIncReasonStarted,
	)
}

func (p *ChainCommitteeLogMetricsProvider) createForChain(chainID isc.ChainID) *ChainCommitteeLogMetrics {
	return newChainCommitteeLogMetrics(p, chainID)
}

type ChainCommitteeLogMetrics struct {
	consOut     prometheus.Counter
	recover     prometheus.Counter
	l1RepAnchor prometheus.Counter
	started     prometheus.Counter
}

func newChainCommitteeLogMetrics(collectors *ChainCommitteeLogMetricsProvider, chainID isc.ChainID) *ChainCommitteeLogMetrics {
	labels := getChainLabels(chainID)
	return &ChainCommitteeLogMetrics{
		consOut:     collectors.logIndexIncReasonConsOut.With(labels),
		recover:     collectors.logIndexIncReasonRecover.With(labels),
		l1RepAnchor: collectors.logIndexIncReasonL1RepAnchor.With(labels),
		started:     collectors.logIndexIncReasonStarted.With(labels),
	}
}

func (m *ChainCommitteeLogMetrics) NextLogIndexCauseConsOut()     { m.consOut.Inc() }
func (m *ChainCommitteeLogMetrics) NextLogIndexCauseRecover()     { m.recover.Inc() }
func (m *ChainCommitteeLogMetrics) NextLogIndexCauseL1RepAnchor() { m.l1RepAnchor.Inc() }
func (m *ChainCommitteeLogMetrics) NextLogIndexCauseStarted()     { m.started.Inc() }
