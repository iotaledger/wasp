package nodeconnmetrics

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type emptyNodeConnectionMessagesMetrics struct{}

func newEmptyNodeConnectionMessagesMetrics() *emptyNodeConnectionMessagesMetrics {
	return &emptyNodeConnectionMessagesMetrics{}
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPublishStateTransaction() NodeConnectionMessageMetrics[*StateTransaction] {
	return newEmptyNodeConnectionMessageMetrics[*StateTransaction]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPublishGovernanceTransaction() NodeConnectionMessageMetrics[*iotago.Transaction] {
	return newEmptyNodeConnectionMessageMetrics[*iotago.Transaction]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPullLatestOutput() NodeConnectionMessageMetrics[interface{}] {
	return newEmptyNodeConnectionMessageMetrics[interface{}]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPullTxInclusionState() NodeConnectionMessageMetrics[iotago.TransactionID] {
	return newEmptyNodeConnectionMessageMetrics[iotago.TransactionID]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPullOutputByID() NodeConnectionMessageMetrics[*iotago.UTXOInput] {
	return newEmptyNodeConnectionMessageMetrics[*iotago.UTXOInput]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInStateOutput() NodeConnectionMessageMetrics[*InStateOutput] {
	return newEmptyNodeConnectionMessageMetrics[*InStateOutput]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInAliasOutput() NodeConnectionMessageMetrics[*iotago.AliasOutput] {
	return newEmptyNodeConnectionMessageMetrics[*iotago.AliasOutput]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInOutput() NodeConnectionMessageMetrics[*InOutput] {
	return newEmptyNodeConnectionMessageMetrics[*InOutput]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInOnLedgerRequest() NodeConnectionMessageMetrics[isc.OnLedgerRequest] {
	return newEmptyNodeConnectionMessageMetrics[isc.OnLedgerRequest]()
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInTxInclusionState() NodeConnectionMessageMetrics[*TxInclusionStateMsg] {
	return newEmptyNodeConnectionMessageMetrics[*TxInclusionStateMsg]()
}
