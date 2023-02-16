package nodeconnmetrics

import "time"

type nodeConnectionMessageRelatedMetricsImpl[T any] struct {
	NodeConnectionMessageMetrics[T]
	related NodeConnectionMessageMetrics[T]
}

func newNodeConnectionMessageRelatedMetrics[T any](metrics, related NodeConnectionMessageMetrics[T]) NodeConnectionMessageMetrics[T] {
	return &nodeConnectionMessageRelatedMetricsImpl[T]{
		NodeConnectionMessageMetrics: metrics,
		related:                      related,
	}
}

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl[T]) GetL1MessagesTotal() uint32 {
	return ncmrmi.NodeConnectionMessageMetrics.GetL1MessagesTotal()
}

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl[T]) GetLastL1MessageTime() time.Time {
	return ncmrmi.NodeConnectionMessageMetrics.GetLastL1MessageTime()
}

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl[T]) GetLastL1Message() T {
	return ncmrmi.NodeConnectionMessageMetrics.GetLastL1Message()
}

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl[T]) IncL1Messages(msg T) {
	ncmrmi.NodeConnectionMessageMetrics.IncL1Messages(msg)
	ncmrmi.related.IncL1Messages(msg)
}
