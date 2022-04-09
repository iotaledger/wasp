package nodeconnmetrics

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/prometheus/client_golang/prometheus"
)

type nodeConnectionMetricsImpl struct {
	NodeConnectionMessagesMetrics
	log                 *logger.Logger
	messageTotalCounter *prometheus.CounterVec
	lastEventTimeGauge  *prometheus.GaugeVec
	registered          []*iscp.ChainID
	inMilestoneMetrics  NodeConnectionMessageMetrics
}

var _ NodeConnectionMetrics = &nodeConnectionMetricsImpl{}

func New(log *logger.Logger) NodeConnectionMetrics {
	ret := &nodeConnectionMetricsImpl{
		log:        log.Named("nodeconn"),
		registered: make([]*iscp.ChainID, 0),
	}
	ret.NodeConnectionMessagesMetrics = newNodeConnectionMessagesMetrics(ret, nil)
	ret.inMilestoneMetrics = newNodeConnectionMessageSimpleMetrics(ret, nil, "in_milestone")
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

func (ncmiT *nodeConnectionMetricsImpl) NewMessagesMetrics(chainID *iscp.ChainID) NodeConnectionMessagesMetrics {
	return newNodeConnectionMessagesMetrics(ncmiT, chainID)
}

// TODO: connect registered to Prometheus
func (ncmiT *nodeConnectionMetricsImpl) SetRegistered(chainID *iscp.ChainID) {
	ncmiT.registered = append(ncmiT.registered, chainID)
}

// TODO: connect registered to Prometheus
func (ncmiT *nodeConnectionMetricsImpl) SetUnregistered(chainID *iscp.ChainID) {
	var i int
	for i = 0; i < len(ncmiT.registered); i++ {
		if ncmiT.registered[i] == chainID {
			ncmiT.registered = append(ncmiT.registered[:i], ncmiT.registered[i+1:]...)
			return
		}
	}
}

func (ncmiT *nodeConnectionMetricsImpl) GetRegistered() []*iscp.ChainID {
	return ncmiT.registered
}

func (ncmiT *nodeConnectionMetricsImpl) GetInMilestone() NodeConnectionMessageMetrics {
	return ncmiT.inMilestoneMetrics
}

func (ncmiT *nodeConnectionMetricsImpl) incTotalPrometheusCounter(label prometheus.Labels) {
	ncmiT.messageTotalCounter.With(label).Inc()
}

func (ncmiT *nodeConnectionMetricsImpl) setLastEventPrometheusGaugeToNow(label prometheus.Labels) {
	ncmiT.lastEventTimeGauge.With(label).SetToCurrentTime()
}
