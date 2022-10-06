package nodeconnmetrics

import (
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
)

type emptyNodeConnectionMetrics struct {
	*emptyNodeConnectionMessagesMetrics
	emptyInMileStoneMetrics NodeConnectionMessageMetrics[*nodeclient.MilestoneInfo]
}

func NewEmptyNodeConnectionMetrics() NodeConnectionMetrics {
	return &emptyNodeConnectionMetrics{}
}

func (encmT *emptyNodeConnectionMetrics) RegisterMetrics() {}

func (encmT *emptyNodeConnectionMetrics) NewMessagesMetrics(chainID *isc.ChainID) NodeConnectionMessagesMetrics {
	encmT.emptyNodeConnectionMessagesMetrics = newEmptyNodeConnectionMessagesMetrics()
	encmT.emptyInMileStoneMetrics = newEmptyNodeConnectionMessageMetrics[*nodeclient.MilestoneInfo]()
	return encmT.emptyNodeConnectionMessagesMetrics
}

func (encmT *emptyNodeConnectionMetrics) SetRegistered(*isc.ChainID)   {}
func (encmT *emptyNodeConnectionMetrics) SetUnregistered(*isc.ChainID) {}

func (encmT *emptyNodeConnectionMetrics) GetRegistered() []*isc.ChainID {
	return []*isc.ChainID{}
}

func (encmT *emptyNodeConnectionMetrics) GetInMilestone() NodeConnectionMessageMetrics[*nodeclient.MilestoneInfo] {
	return encmT.emptyInMileStoneMetrics
}
