package nodeconnmetrics

import (
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
)

type nodeConnectionMetricsImpl struct {
	NodeConnectionMessagesMetrics
	log                 *logger.Logger
	messageTotalCounter *prometheus.CounterVec
	lastEventTimeGauge  *prometheus.GaugeVec
	registered          []*isc.ChainID
	inMilestoneMetrics  NodeConnectionMessageMetrics[*nodeclient.MilestoneInfo]
}

var _ NodeConnectionMetrics = &nodeConnectionMetricsImpl{}

func New(log *logger.Logger) NodeConnectionMetrics {
	ret := &nodeConnectionMetricsImpl{
		log:        log.Named("nodeconn"),
		registered: make([]*isc.ChainID, 0),
	}
	ret.NodeConnectionMessagesMetrics = newNodeConnectionMessagesMetrics(ret, nil)
	ret.inMilestoneMetrics = newNodeConnectionMessageSimpleMetrics[*nodeclient.MilestoneInfo](ret, nil, "in_milestone")
	return ret
}

func (ncmiT *nodeConnectionMetricsImpl) RegisterMetrics() {
	ncmiT.log.Debug("Registering nodeconnection metrics to prometheus...")
	ncmiT.messageTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_nodeconn_message_total_counter",
		Help: "Number of messages send/received by node connection of the chain",
	}, []string{chainLabelNameConst, msgTypeLabelNameConst})
	prometheus.MustRegister(ncmiT.messageTotalCounter)
	ncmiT.lastEventTimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "wasp_nodeconn_last_event_time_gauge",
		Help: "Last time when the message was sent/received by node connection of the chain",
	}, []string{chainLabelNameConst, msgTypeLabelNameConst})
	prometheus.MustRegister(ncmiT.lastEventTimeGauge)
	ncmiT.log.Info("Registering nodeconnection metrics to prometheus... Done")
}

func (ncmiT *nodeConnectionMetricsImpl) NewMessagesMetrics(chainID *isc.ChainID) NodeConnectionMessagesMetrics {
	return newNodeConnectionMessagesMetrics(ncmiT, chainID)
}

// TODO: connect registered to Prometheus
func (ncmiT *nodeConnectionMetricsImpl) SetRegistered(chainID *isc.ChainID) {
	ncmiT.registered = append(ncmiT.registered, chainID)
}

// TODO: connect registered to Prometheus
func (ncmiT *nodeConnectionMetricsImpl) SetUnregistered(chainID *isc.ChainID) {
	var i int
	for i = 0; i < len(ncmiT.registered); i++ {
		if ncmiT.registered[i] == chainID {
			ncmiT.registered = append(ncmiT.registered[:i], ncmiT.registered[i+1:]...)
			return
		}
	}
}

func (ncmiT *nodeConnectionMetricsImpl) GetRegistered() []*isc.ChainID {
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
