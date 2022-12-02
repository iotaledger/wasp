package nodeconnmetrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/iota.go/v3/nodeclient"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type StateTransaction struct {
	StateIndex  uint32
	Transaction *iotago.Transaction
}

type InStateOutput struct {
	OutputID iotago.OutputID
	Output   iotago.Output
}

type InOutput struct {
	OutputID iotago.OutputID
	Output   iotago.Output
}

type TxInclusionStateMsg struct {
	TxID  iotago.TransactionID
	State string
}

type NodeConnectionMessageMetrics[T any] interface {
	CountLastMessage(T)
	GetMessageTotal() uint32
	GetLastEvent() time.Time
	GetLastMessage() T
}

type NodeConnectionMessagesMetrics interface {
	GetOutPublishStateTransaction() NodeConnectionMessageMetrics[*StateTransaction]
	GetOutPublishGovernanceTransaction() NodeConnectionMessageMetrics[*iotago.Transaction]
	GetOutPullLatestOutput() NodeConnectionMessageMetrics[interface{}]
	GetOutPullTxInclusionState() NodeConnectionMessageMetrics[iotago.TransactionID]
	GetOutPullOutputByID() NodeConnectionMessageMetrics[*iotago.UTXOInput]
	GetInStateOutput() NodeConnectionMessageMetrics[*InStateOutput]
	GetInAliasOutput() NodeConnectionMessageMetrics[*iotago.AliasOutput]
	GetInOutput() NodeConnectionMessageMetrics[*InOutput]
	GetInOnLedgerRequest() NodeConnectionMessageMetrics[isc.OnLedgerRequest]
	GetInTxInclusionState() NodeConnectionMessageMetrics[*TxInclusionStateMsg]
}

type NodeConnectionMetrics interface {
	NodeConnectionMessagesMetrics
	GetInMilestone() NodeConnectionMessageMetrics[*nodeclient.MilestoneInfo]
	SetRegistered(*isc.ChainID)
	SetUnregistered(*isc.ChainID)
	GetRegistered() []*isc.ChainID
	Register(registry *prometheus.Registry)
	NewMessagesMetrics(*isc.ChainID) NodeConnectionMessagesMetrics
}
