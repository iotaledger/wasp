package nodeconnmetrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
)

type nodeConnectionMetricsImpl struct {
	NodeConnectionMessagesMetrics
	registered []isc.ChainID

	messageTotalCounter *prometheus.CounterVec
	lastEventTimeGauge  *prometheus.GaugeVec

	inMilestoneMetrics NodeConnectionMessageMetrics[*nodeclient.MilestoneInfo]
}

var _ NodeConnectionMetrics = &nodeConnectionMetricsImpl{}

func New() NodeConnectionMetrics {
	ret := &nodeConnectionMetricsImpl{
		registered: make([]isc.ChainID, 0),

		messageTotalCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota",
			Subsystem: "wasp_nodeconn",
			Name:      "message_total_counter",
			Help:      "Number of messages send/received by node connection of the chain",
		}, []string{chainLabelNameConst, msgTypeLabelNameConst}),

		lastEventTimeGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota",
			Subsystem: "wasp_nodeconn",
			Name:      "last_event_time_gauge",
			Help:      "Last time when the message was sent/received by node connection of the chain",
		}, []string{chainLabelNameConst, msgTypeLabelNameConst}),
	}
	ret.NodeConnectionMessagesMetrics = newNodeConnectionMessagesMetrics(ret, isc.ChainID{})
	ret.inMilestoneMetrics = newNodeConnectionMessageSimpleMetrics[*nodeclient.MilestoneInfo](ret, isc.ChainID{}, "in_milestone")
	return ret
}

func (ncmiT *nodeConnectionMetricsImpl) Register(registry *prometheus.Registry) {
	registry.MustRegister(
		ncmiT.messageTotalCounter,
		ncmiT.lastEventTimeGauge,
	)
}

func (ncmiT *nodeConnectionMetricsImpl) NewMessagesMetrics(chainID isc.ChainID) NodeConnectionMessagesMetrics {
	return newNodeConnectionMessagesMetrics(ncmiT, chainID)
}

// TODO: connect registered to Prometheus
func (ncmiT *nodeConnectionMetricsImpl) SetRegistered(chainID isc.ChainID) {
	ncmiT.registered = append(ncmiT.registered, chainID)
}

// TODO: connect registered to Prometheus
func (ncmiT *nodeConnectionMetricsImpl) SetUnregistered(chainID isc.ChainID) {
	var i int
	for i = 0; i < len(ncmiT.registered); i++ {
		if ncmiT.registered[i] == chainID {
			ncmiT.registered = append(ncmiT.registered[:i], ncmiT.registered[i+1:]...)
			return
		}
	}
}

func (ncmiT *nodeConnectionMetricsImpl) GetRegistered() []isc.ChainID {
	return ncmiT.registered
}

func (ncmiT *nodeConnectionMetricsImpl) GetInMilestone() NodeConnectionMessageMetrics[*nodeclient.MilestoneInfo] {
	return ncmiT.inMilestoneMetrics
}

func (ncmiT *nodeConnectionMetricsImpl) incTotalPrometheusCounter(label prometheus.Labels) {
	ncmiT.messageTotalCounter.With(label).Inc()
}

func (ncmiT *nodeConnectionMetricsImpl) setLastEventPrometheusGaugeToNow(label prometheus.Labels) {
	ncmiT.lastEventTimeGauge.With(label).SetToCurrentTime()
}
