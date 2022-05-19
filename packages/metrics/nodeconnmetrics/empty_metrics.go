package nodeconnmetrics

import (
	"github.com/iotaledger/wasp/packages/iscp"
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

func (encmT *emptyNodeConnectionMetrics) NewMessagesMetrics(chainID *iscp.ChainID) NodeConnectionMessagesMetrics {
	return newEmptyNodeConnectionMessagesMetrics()
}

func (encmT *emptyNodeConnectionMetrics) SetRegistered(*iscp.ChainID)   {}
func (encmT *emptyNodeConnectionMetrics) SetUnregistered(*iscp.ChainID) {}

func (encmT *emptyNodeConnectionMetrics) GetRegistered() []*iscp.ChainID {
	return []*iscp.ChainID{}
}

func (encmT *emptyNodeConnectionMetrics) GetInMilestone() NodeConnectionMessageMetrics {
	return encmT.emptyMessageMetrics
}
