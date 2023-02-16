package nodeconnmetrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
)

type nodeConnectionMetricsImpl struct {
	NodeConnectionMessagesMetrics
	registered []isc.ChainID

	messagesL1        *prometheus.CounterVec
	lastL1MessageTime *prometheus.GaugeVec

	inMilestoneMetrics NodeConnectionMessageMetrics[*nodeclient.MilestoneInfo]
}

var _ NodeConnectionMetrics = &nodeConnectionMetricsImpl{}

func New() NodeConnectionMetrics {
	ret := &nodeConnectionMetricsImpl{
		registered: make([]isc.ChainID, 0),
		messagesL1: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iota_wasp",
			Subsystem: "l1",
			Name:      "messages_total",
			Help:      "Number of messages sent/received by L1 connection of the chain",
		}, []string{chainLabelNameConst, msgTypeLabelNameConst}),

		lastL1MessageTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "iota_wasp",
			Subsystem: "l1",
			Name:      "last_message_time",
			Help:      "Last time when a message was sent/received by L1 connection of the chain",
		}, []string{chainLabelNameConst, msgTypeLabelNameConst}),
	}
	ret.NodeConnectionMessagesMetrics = newNodeConnectionMessagesMetrics(ret, isc.ChainID{})
	ret.inMilestoneMetrics = newNodeConnectionMessageSimpleMetrics[*nodeclient.MilestoneInfo](ret, isc.ChainID{}, "in_milestone")
	return ret
}

func (ncmiT *nodeConnectionMetricsImpl) Register(registry *prometheus.Registry) {
	registry.MustRegister(
		ncmiT.messagesL1,
		ncmiT.lastL1MessageTime,
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
	for i := 0; i < len(ncmiT.registered); i++ {
		if ncmiT.registered[i] == chainID {
			// remove the found chain from the slice and return
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

func (ncmiT *nodeConnectionMetricsImpl) incL1Messages(label prometheus.Labels) {
	ncmiT.messagesL1.With(label).Inc()
}

func (ncmiT *nodeConnectionMetricsImpl) setLastL1MessageTimeToNow(label prometheus.Labels) {
	ncmiT.lastL1MessageTime.With(label).SetToCurrentTime()
}
