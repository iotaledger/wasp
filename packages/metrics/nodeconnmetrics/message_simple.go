package nodeconnmetrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"

	"github.com/iotaledger/wasp/packages/isc"
)

type nodeConnectionMessageSimpleMetrics[T any] struct {
	nodeConnMetrics *nodeConnectionMetricsImpl
	metricsLabel    prometheus.Labels
	total           atomic.Uint32
	lastEvent       time.Time
	lastMessage     T
}

func newNodeConnectionMessageSimpleMetrics[T any](ncmi *nodeConnectionMetricsImpl, chainID *isc.ChainID, msgType string) NodeConnectionMessageMetrics[T] {
	return &nodeConnectionMessageSimpleMetrics[T]{
		nodeConnMetrics: ncmi,
		metricsLabel:    getMetricsLabel(chainID, msgType),
	}
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) incMessageTotal() {
	ncmmi.total.Inc()
	ncmmi.nodeConnMetrics.incTotalPrometheusCounter(ncmmi.metricsLabel)
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) setLastEventToNow() {
	ncmmi.lastEvent = time.Now()
	ncmmi.nodeConnMetrics.setLastEventPrometheusGaugeToNow(ncmmi.metricsLabel)
}

// TODO: connect last message to Prometheus
func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) setLastMessage(msg T) {
	ncmmi.lastMessage = msg
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) CountLastMessage(msg T) {
	ncmmi.incMessageTotal()
	ncmmi.setLastEventToNow()
	ncmmi.setLastMessage(msg)
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) GetMessageTotal() uint32 {
	return ncmmi.total.Load()
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) GetLastEvent() time.Time {
	return ncmmi.lastEvent
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) GetLastMessage() T {
	return ncmmi.lastMessage
}
