package nodeconnmetrics

import (
	iotago "github.com/iotaledger/iota.go/v3"
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
func (ncmi *emptyNodeConnectionMetrics) SetSubscribed(address iotago.Address)   {}
func (ncmi *emptyNodeConnectionMetrics) SetUnsubscribed(address iotago.Address) {}
func (ncmi *emptyNodeConnectionMetrics) GetSubscribed() []iotago.Address {
	return []iotago.Address{}
}
