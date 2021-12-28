package nodeconnmetrics

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"
)

type nodeConnectionMessageMetricsImpl struct {
	nodeConnMetrics *nodeConnectionMetricsImpl
	metricsLabel    prometheus.Labels
	total           atomic.Uint32
	lastEvent       time.Time
	lastMessage     interface{}
}

var _ NodeConnectionMessageMetrics = &nodeConnectionMessageMetricsImpl{}

func newNodeConnectionMessageMetrics(ncmi *nodeConnectionMetricsImpl, chainID *iscp.ChainID, msgType string) NodeConnectionMessageMetrics {
	return &nodeConnectionMessageMetricsImpl{
		nodeConnMetrics: ncmi,
		metricsLabel:    ncmi.getMetricsLabel(chainID, msgType),
	}
}

func (ncmmi *nodeConnectionMessageMetricsImpl) incMessageTotal() {
	ncmmi.total.Inc()
	ncmmi.nodeConnMetrics.messageTotalCounter.With(ncmmi.metricsLabel).Inc()
}

func (ncmmi *nodeConnectionMessageMetricsImpl) setLastEventToNow() {
	ncmmi.lastEvent = time.Now()
	ncmmi.nodeConnMetrics.lastEventTimeGauge.With(ncmmi.metricsLabel).SetToCurrentTime()
}

// TODO: connect last message to Prometheus
func (ncmmi *nodeConnectionMessageMetricsImpl) setLastMessage(msg interface{}) {
	ncmmi.lastMessage = msg
}

func (ncmmi *nodeConnectionMessageMetricsImpl) CountLastMessage(msg interface{}) {
	ncmmi.incMessageTotal()
	ncmmi.setLastEventToNow()
	ncmmi.setLastMessage(msg)
}

func (ncmmi *nodeConnectionMessageMetricsImpl) GetMessageTotal() uint32 {
	return ncmmi.total.Load()
}

func (ncmmi *nodeConnectionMessageMetricsImpl) GetLastEvent() time.Time {
	return ncmmi.lastEvent
}

func (ncmmi *nodeConnectionMessageMetricsImpl) GetLastMessage() interface{} {
	return ncmmi.lastMessage
}
