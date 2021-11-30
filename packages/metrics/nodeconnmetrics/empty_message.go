package nodeconnmetrics

import (
	"time"
)

type emptyNodeConnectionMessageMetrics struct{}

var _ NodeConnectionMessageMetrics = &emptyNodeConnectionMessageMetrics{}

func newEmptyNodeConnectionMessageMetrics() NodeConnectionMessageMetrics {
	return &emptyNodeConnectionMessageMetrics{}
}

func (ncmmi *emptyNodeConnectionMessageMetrics) CountLastMessage(msg interface{}) {}
func (ncmmi *emptyNodeConnectionMessageMetrics) GetMessageTotal() uint32          { return 0 }
func (ncmmi *emptyNodeConnectionMessageMetrics) GetLastEvent() time.Time          { return time.Time{} }
func (ncmmi *emptyNodeConnectionMessageMetrics) GetLastMessage() interface{}      { return nil }
