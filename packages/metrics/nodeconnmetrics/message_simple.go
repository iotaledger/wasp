package nodeconnmetrics

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"
)

type nodeConnectionMessageSimpleMetrics struct {
	nodeConnMetrics *nodeConnectionMetricsImpl
	metricsLabel    prometheus.Labels
	total           atomic.Uint32
	lastEvent       time.Time
	lastMessage     interface{}
}

var _ NodeConnectionMessageMetrics = &nodeConnectionMessageSimpleMetrics{}

func newNodeConnectionMessageSimpleMetrics(ncmi *nodeConnectionMetricsImpl, chainID *iscp.ChainID, msgType string) NodeConnectionMessageMetrics {
	return &nodeConnectionMessageSimpleMetrics{
		nodeConnMetrics: ncmi,
		metricsLabel:    getMetricsLabel(chainID, msgType),
	}
}

func (ncmmi *nodeConnectionMessageSimpleMetrics) incMessageTotal() {
	ncmmi.total.Inc()
	ncmmi.nodeConnMetrics.incTotalPrometheusCounter(ncmmi.metricsLabel)
}

func (ncmmi *nodeConnectionMessageSimpleMetrics) setLastEventToNow() {
	ncmmi.lastEvent = time.Now()
	ncmmi.nodeConnMetrics.setLastEventPrometheusGaugeToNow(ncmmi.metricsLabel)
}

// TODO: connect last message to Prometheus
func (ncmmi *nodeConnectionMessageSimpleMetrics) setLastMessage(msg interface{}) {
	ncmmi.lastMessage = msg
}

func (ncmmi *nodeConnectionMessageSimpleMetrics) CountLastMessage(msg interface{}) {
	ncmmi.incMessageTotal()
	ncmmi.setLastEventToNow()
	ncmmi.setLastMessage(msg)
}

func (ncmmi *nodeConnectionMessageSimpleMetrics) GetMessageTotal() uint32 {
	return ncmmi.total.Load()
}

func (ncmmi *nodeConnectionMessageSimpleMetrics) GetLastEvent() time.Time {
	return ncmmi.lastEvent
}

func (ncmmi *nodeConnectionMessageSimpleMetrics) GetLastMessage() interface{} {
	return ncmmi.lastMessage
}
