package nodeconnmetrics

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
)

type NodeConnectionMessageMetrics interface {
	CountLastMessage(interface{})
	GetMessageTotal() uint32
	GetLastEvent() time.Time
	GetLastMessage() interface{}
}

type NodeConnectionMessagesMetrics interface {
	GetOutPublishTransaction() NodeConnectionMessageMetrics
	GetOutPullLatestOutput() NodeConnectionMessageMetrics
	GetOutPullTxInclusionState() NodeConnectionMessageMetrics
	GetOutPullOutputByID() NodeConnectionMessageMetrics
	GetInOutput() NodeConnectionMessageMetrics
	GetInAliasOutput() NodeConnectionMessageMetrics
	GetInOnLedgerRequest() NodeConnectionMessageMetrics
	GetInTxInclusionState() NodeConnectionMessageMetrics
}

type NodeConnectionMetrics interface {
	NodeConnectionMessagesMetrics
	GetInMilestone() NodeConnectionMessageMetrics
	SetRegistered(*iscp.ChainID)
	SetUnregistered(*iscp.ChainID)
	GetRegistered() []*iscp.ChainID
	RegisterMetrics()
	NewMessagesMetrics(*iscp.ChainID) NodeConnectionMessagesMetrics
}
