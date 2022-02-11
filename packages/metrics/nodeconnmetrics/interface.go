package nodeconnmetrics

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
)

type NodeConnectionMessageMetrics interface {
	CountLastMessage(interface{})
	GetMessageTotal() uint32
	GetLastEvent() time.Time
	GetLastMessage() interface{}
}

type NodeConnectionMessagesMetrics interface {
	GetOutPullState() NodeConnectionMessageMetrics
	GetOutPullTransactionInclusionState() NodeConnectionMessageMetrics
	GetOutPullConfirmedOutput() NodeConnectionMessageMetrics
	GetOutPostTransaction() NodeConnectionMessageMetrics
	GetInTransaction() NodeConnectionMessageMetrics
	GetInInclusionState() NodeConnectionMessageMetrics
	GetInOutput() NodeConnectionMessageMetrics
	GetInUnspentAliasOutput() NodeConnectionMessageMetrics
}

type NodeConnectionMetrics interface {
	NodeConnectionMessagesMetrics
	SetSubscribed(iotago.Address)
	SetUnsubscribed(iotago.Address)
	GetSubscribed() []iotago.Address
	RegisterMetrics()
	NewMessagesMetrics(chainID *iscp.ChainID) NodeConnectionMessagesMetrics
}
