package nodeconnmetrics

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
)

type emptyNodeConnectionMetrics struct {
	NodeConnectionMessagesMetrics
}

var _ NodeConnectionMetrics = &emptyNodeConnectionMetrics{}

func NewEmptyNodeConnectionMetrics() NodeConnectionMetrics {
	return &emptyNodeConnectionMetrics{
		NodeConnectionMessagesMetrics: NewEmptyNodeConnectionMessagesMetrics(),
	}
}

func (ncmi *emptyNodeConnectionMetrics) RegisterMetrics() {}
func (ncmi *emptyNodeConnectionMetrics) NewMessagesMetrics(chainID *iscp.ChainID) NodeConnectionMessagesMetrics {
	return NewEmptyNodeConnectionMessagesMetrics()
}
func (ncmi *emptyNodeConnectionMetrics) SetSubscribed(address ledgerstate.Address)   {}
func (ncmi *emptyNodeConnectionMetrics) SetUnsubscribed(address ledgerstate.Address) {}
func (ncmi *emptyNodeConnectionMetrics) GetSubscribed() []ledgerstate.Address {
	return []ledgerstate.Address{}
}
