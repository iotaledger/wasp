package nodeconnmetrics

import (
	"time"

	"github.com/iotaledger/wasp/packages/isc"
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
	SetRegistered(*isc.ChainID)
	SetUnregistered(*isc.ChainID)
	GetRegistered() []*isc.ChainID
	RegisterMetrics()
	NewMessagesMetrics(*isc.ChainID) NodeConnectionMessagesMetrics
}
