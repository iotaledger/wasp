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
	GetOutPublishStateTransaction() NodeConnectionMessageMetrics
	GetOutPublishGovernanceTransaction() NodeConnectionMessageMetrics
	GetOutPullLatestOutput() NodeConnectionMessageMetrics
	GetOutPullTxInclusionState() NodeConnectionMessageMetrics
	GetOutPullOutputByID() NodeConnectionMessageMetrics
	GetInStateOutput() NodeConnectionMessageMetrics
	GetInAliasOutput() NodeConnectionMessageMetrics
	GetInOutput() NodeConnectionMessageMetrics
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
