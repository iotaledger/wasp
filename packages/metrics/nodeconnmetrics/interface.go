package nodeconnmetrics

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
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
	SetRegistered(iotago.Address)
	SetUnregistered(iotago.Address)
	GetRegistered() []iotago.Address
	RegisterMetrics()
	NewMessagesMetrics(iotago.Address) NodeConnectionMessagesMetrics
}
