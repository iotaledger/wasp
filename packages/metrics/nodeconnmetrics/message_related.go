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

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl[T]) GetMessageTotal() uint32 {
	return ncmrmi.NodeConnectionMessageMetrics.GetMessageTotal()
}

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl[T]) GetLastEvent() time.Time {
	return ncmrmi.NodeConnectionMessageMetrics.GetLastEvent()
}

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl[T]) GetLastMessage() T {
	return ncmrmi.NodeConnectionMessageMetrics.GetLastMessage()
}

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl[T]) CountLastMessage(msg T) {
	ncmrmi.NodeConnectionMessageMetrics.CountLastMessage(msg)
	ncmrmi.related.CountLastMessage(msg)
}
