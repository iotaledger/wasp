package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

type ChainCmtLogMetricsProvider struct {
	logIndexIncReasonConsOut *prometheus.CounterVec
	logIndexIncReasonRecover *prometheus.CounterVec
	logIndexIncReasonL1RepAO *prometheus.CounterVec
	logIndexIncReasonStarted *prometheus.CounterVec
}

func newChainCmtLogMetricsProvider() *ChainCmtLogMetricsProvider {
	return &ChainCmtLogMetricsProvider{
		logIndexIncReasonConsOut: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "cmt_log",
			Name:      "log_index_inc_reason_ConsOut",
			Help:      "Number if times LogIndex was increased because of consensus output.",
		}, []string{labelNameChain}),
		logIndexIncReasonRecover: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "cmt_log",
			Name:      "log_index_inc_reason_Recover",
			Help:      "Number if times LogIndex was increased because of recovery procedure.",
		}, []string{labelNameChain}),
		logIndexIncReasonL1RepAO: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "cmt_log",
			Name:      "log_index_inc_reason_L1RepAO",
			Help:      "Number if times LogIndex was increased because L1 replaced TIP alias output.",
		}, []string{labelNameChain}),
		logIndexIncReasonStarted: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "cmt_log",
			Name:      "log_index_inc_reason_Started",
			Help:      "Number if times LogIndex was increased because other nodes started the consensus.",
		}, []string{labelNameChain}),
	}
}

func (p *ChainCmtLogMetricsProvider) register(reg prometheus.Registerer) {
	reg.MustRegister(
		p.logIndexIncReasonConsOut,
		p.logIndexIncReasonRecover,
		p.logIndexIncReasonL1RepAO,
		p.logIndexIncReasonStarted,
	)
}

func (p *ChainCmtLogMetricsProvider) createForChain(chainID isc.ChainID) *ChainCmtLogMetrics {
	return newChainCmtLogMetrics(p, chainID)
}

type ChainCmtLogMetrics struct {
	consOut prometheus.Counter
	recover prometheus.Counter
	l1RepAO prometheus.Counter
	started prometheus.Counter
}

func newChainCmtLogMetrics(collectors *ChainCmtLogMetricsProvider, chainID isc.ChainID) *ChainCmtLogMetrics {
	labels := getChainLabels(chainID)
	return &ChainCmtLogMetrics{
		consOut: collectors.logIndexIncReasonConsOut.With(labels),
		recover: collectors.logIndexIncReasonRecover.With(labels),
		l1RepAO: collectors.logIndexIncReasonL1RepAO.With(labels),
		started: collectors.logIndexIncReasonStarted.With(labels),
	}
}

func (m *ChainCmtLogMetrics) NextLogIndexCauseConsOut() { m.consOut.Inc() }
func (m *ChainCmtLogMetrics) NextLogIndexCauseRecover() { m.recover.Inc() }
func (m *ChainCmtLogMetrics) NextLogIndexCauseL1RepAO() { m.l1RepAO.Inc() }
func (m *ChainCmtLogMetrics) NextLogIndexCauseStarted() { m.started.Inc() }
