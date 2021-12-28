package nodeconnmetrics

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
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

	SetSubscribed(ledgerstate.Address)
	SetUnsubscribed(ledgerstate.Address)
	GetSubscribed() []ledgerstate.Address

	RegisterMetrics()
	NewMessagesMetrics(chainID *iscp.ChainID) NodeConnectionMessagesMetrics
}
