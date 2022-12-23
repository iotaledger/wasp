package nodeconnmetrics

import (
	"time"
)

type emptyNodeConnectionMessageMetrics[T any] struct{}

func newEmptyNodeConnectionMessageMetrics[T any]() NodeConnectionMessageMetrics[T] {
	return &emptyNodeConnectionMessageMetrics[T]{}
}

func (ncmmi *emptyNodeConnectionMessageMetrics[T]) CountLastMessage(msg T)  {}
func (ncmmi *emptyNodeConnectionMessageMetrics[T]) GetMessageTotal() uint32 { return 0 }
func (ncmmi *emptyNodeConnectionMessageMetrics[T]) GetLastEvent() time.Time { return time.Time{} }
func (ncmmi *emptyNodeConnectionMessageMetrics[T]) GetLastMessage() T {
	var result T
	return result
}
