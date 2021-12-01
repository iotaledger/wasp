package nodeconnmetrics

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/prometheus/client_golang/prometheus"
)

type nodeConnectionMetricsImpl struct {
	NodeConnectionMessagesMetrics
	log                 *logger.Logger
	messageTotalCounter *prometheus.CounterVec
	lastEventTimeGauge  *prometheus.GaugeVec
	subscribed          []ledgerstate.Address
}

var _ NodeConnectionMetrics = &nodeConnectionMetricsImpl{}

const (
	chainLabelName   = "chain"
	msgTypeLabelName = "message_type"
)

func New(log *logger.Logger) NodeConnectionMetrics {
	ret := &nodeConnectionMetricsImpl{log: log.Named("nodeconn")}
	ret.NodeConnectionMessagesMetrics = newNodeConnectionMessagesMetrics(ret, nil)
	return ret
}

func (ncmi *nodeConnectionMetricsImpl) RegisterMetrics() {
	ncmi.log.Debug("Registering nodeconnection metrics to prometheus...")
	ncmi.messageTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wasp_nodeconn_message_total_counter",
		Help: "Number of messages send/received by node connection of the chain",
	}, []string{chainLabelName, msgTypeLabelName})
	prometheus.MustRegister(ncmi.messageTotalCounter)
	ncmi.lastEventTimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "wasp_nodeconn_last_event_time_gauge",
		Help: "Last time when the message was sent/received by node connection of the chain",
	}, []string{chainLabelName, msgTypeLabelName})
	prometheus.MustRegister(ncmi.lastEventTimeGauge)
	ncmi.log.Info("Registering nodeconnection metrics to prometheus... Done")
}

func (ncmi *nodeConnectionMetricsImpl) NewMessagesMetrics(chainID *iscp.ChainID) NodeConnectionMessagesMetrics {
	return newNodeConnectionMessagesMetrics(ncmi, chainID)
}

func (ncmi *nodeConnectionMetricsImpl) getMetricsLabel(chainID *iscp.ChainID, msgType string) prometheus.Labels {
	var chainIDStr string
	if chainID == nil {
		chainIDStr = ""
	} else {
		chainIDStr = chainID.String()
	}
	return prometheus.Labels{
		chainLabelName:   chainIDStr,
		msgTypeLabelName: msgType,
	}
}

// TODO: connect subscribed to Prometheus
func (ncmi *nodeConnectionMetricsImpl) SetSubscribed(address ledgerstate.Address) {
	ncmi.subscribed = append(ncmi.subscribed, address)
}

// TODO: connect subscribed to Prometheus
func (ncmi *nodeConnectionMetricsImpl) SetUnsubscribed(address ledgerstate.Address) {
	var i int
	for i = 0; i < len(ncmi.subscribed); i++ {
		if ncmi.subscribed[i] == address {
			ncmi.subscribed = append(ncmi.subscribed[:i], ncmi.subscribed[i+1:]...)
			return
		}
	}
}

func (ncmi *nodeConnectionMetricsImpl) GetSubscribed() []ledgerstate.Address {
	return ncmi.subscribed
}
