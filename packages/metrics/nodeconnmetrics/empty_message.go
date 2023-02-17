package nodeconnmetrics

import (
	"time"
)

type emptyNodeConnectionMessageMetrics[T any] struct{}

func newEmptyNodeConnectionMessageMetrics[T any]() NodeConnectionMessageMetrics[T] {
	return &emptyNodeConnectionMessageMetrics[T]{}
}

func (ncmmi *emptyNodeConnectionMessageMetrics[T]) IncL1Messages(msg T)        {}
func (ncmmi *emptyNodeConnectionMessageMetrics[T]) GetL1MessagesTotal() uint32 { return 0 }
func (ncmmi *emptyNodeConnectionMessageMetrics[T]) GetLastL1MessageTime() time.Time {
	return time.Time{}
}

func (ncmmi *emptyNodeConnectionMessageMetrics[T]) GetLastL1Message() T {
	var result T
	return result
}
