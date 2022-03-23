package nodeconnmetrics

import (
	iotago "github.com/iotaledger/iota.go/v3"
)

type emptyNodeConnectionMetrics struct {
	*emptyNodeConnectionMessagesMetrics
	emptyMessageMetrics NodeConnectionMessageMetrics
}

var _ NodeConnectionMetrics = &emptyNodeConnectionMetrics{}

func NewEmptyNodeConnectionMetrics() NodeConnectionMetrics {
	return &emptyNodeConnectionMetrics{
		emptyNodeConnectionMessagesMetrics: newEmptyNodeConnectionMessagesMetrics(),
		emptyMessageMetrics:                newEmptyNodeConnectionMessageMetrics(),
	}
}

func (encmT *emptyNodeConnectionMetrics) RegisterMetrics() {}

func (encmT *emptyNodeConnectionMetrics) NewMessagesMetrics(chainAddr iotago.Address) NodeConnectionMessagesMetrics {
	return newEmptyNodeConnectionMessagesMetrics()
}

func (encmT *emptyNodeConnectionMetrics) SetRegistered(iotago.Address)   {}
func (encmT *emptyNodeConnectionMetrics) SetUnregistered(iotago.Address) {}

func (encmT *emptyNodeConnectionMetrics) GetRegistered() []iotago.Address {
	return []iotago.Address{}
}

func (encmT *emptyNodeConnectionMetrics) GetInMilestone() NodeConnectionMessageMetrics {
	return encmT.emptyMessageMetrics
}
