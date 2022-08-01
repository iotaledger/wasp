package nodeconnmetrics

import (
	"github.com/iotaledger/wasp/packages/isc"
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

func (encmT *emptyNodeConnectionMetrics) NewMessagesMetrics(chainID *isc.ChainID) NodeConnectionMessagesMetrics {
	return newEmptyNodeConnectionMessagesMetrics()
}

func (encmT *emptyNodeConnectionMetrics) SetRegistered(*isc.ChainID)   {}
func (encmT *emptyNodeConnectionMetrics) SetUnregistered(*isc.ChainID) {}

func (encmT *emptyNodeConnectionMetrics) GetRegistered() []*isc.ChainID {
	return []*isc.ChainID{}
}

func (encmT *emptyNodeConnectionMetrics) GetInMilestone() NodeConnectionMessageMetrics {
	return encmT.emptyMessageMetrics
}
