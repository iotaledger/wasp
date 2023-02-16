package nodeconnmetrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"

	"github.com/iotaledger/wasp/packages/isc"
)

type nodeConnectionMessageSimpleMetrics[T any] struct {
	nodeConnMetrics   *nodeConnectionMetricsImpl
	metricsLabel      prometheus.Labels
	messagesL1        atomic.Uint32
	lastL1MessageTime time.Time
	lastL1Message     T
}

func newNodeConnectionMessageSimpleMetrics[T any](ncmi *nodeConnectionMetricsImpl, chainID isc.ChainID, msgType string) NodeConnectionMessageMetrics[T] {
	return &nodeConnectionMessageSimpleMetrics[T]{
		nodeConnMetrics: ncmi,
		metricsLabel:    getMetricsLabel(chainID, msgType),
	}
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) incL1Messages() {
	ncmmi.messagesL1.Inc()
	ncmmi.nodeConnMetrics.incL1Messages(ncmmi.metricsLabel)
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) setLastL1MessageTimeToNow() {
	ncmmi.lastL1MessageTime = time.Now()
	ncmmi.nodeConnMetrics.setLastL1MessageTimeToNow(ncmmi.metricsLabel)
}

// TODO: connect last message to Prometheus
func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) setLastL1Message(msg T) {
	ncmmi.lastL1Message = msg
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) IncL1Messages(msg T) {
	ncmmi.incL1Messages()
	ncmmi.setLastL1MessageTimeToNow()
	ncmmi.setLastL1Message(msg)
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) GetL1MessagesTotal() uint32 {
	return ncmmi.messagesL1.Load()
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) GetLastL1MessageTime() time.Time {
	return ncmmi.lastL1MessageTime
}

func (ncmmi *nodeConnectionMessageSimpleMetrics[T]) GetLastL1Message() T {
	return ncmmi.lastL1Message
}
