package nodeconnmetrics

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type nodeConnectionMessagesMetricsImpl struct {
	outPublishStateTransactionMetrics      NodeConnectionMessageMetrics[*StateTransaction]
	outPublishGovernanceTransactionMetrics NodeConnectionMessageMetrics[*iotago.Transaction]
	outPullLatestOutputMetrics             NodeConnectionMessageMetrics[interface{}]
	outPullTxInclusionStateMetrics         NodeConnectionMessageMetrics[iotago.TransactionID]
	outPullOutputByIDMetrics               NodeConnectionMessageMetrics[iotago.OutputID]

	inStateOutputMetrics      NodeConnectionMessageMetrics[*InStateOutput]
	inAliasOutputMetrics      NodeConnectionMessageMetrics[*iotago.AliasOutput]
	inOutputMetrics           NodeConnectionMessageMetrics[*InOutput]
	inOnLedgerRequestMetrics  NodeConnectionMessageMetrics[isc.OnLedgerRequest]
	inTxInclusionStateMetrics NodeConnectionMessageMetrics[*TxInclusionStateMsg]
}

var _ NodeConnectionMessagesMetrics = &nodeConnectionMessagesMetricsImpl{}

func createMetricsMessage[T any](ncmi *nodeConnectionMetricsImpl, chainID *isc.ChainID, msgType string, makeRelatedMetricsFun func() NodeConnectionMessageMetrics[T]) NodeConnectionMessageMetrics[T] {
	simpleMessageMetrics := newNodeConnectionMessageSimpleMetrics[T](ncmi, chainID, msgType)
	if chainID == nil {
		return simpleMessageMetrics
	}

	return newNodeConnectionMessageRelatedMetrics(simpleMessageMetrics, makeRelatedMetricsFun())
}

func newNodeConnectionMessagesMetrics(ncmi *nodeConnectionMetricsImpl, chainID *isc.ChainID) NodeConnectionMessagesMetrics {
	return &nodeConnectionMessagesMetricsImpl{
		outPublishStateTransactionMetrics: createMetricsMessage(ncmi, chainID, "out_publish_state_transaction", func() NodeConnectionMessageMetrics[*StateTransaction] {
			return ncmi.GetOutPublishStateTransaction()
		}),
		outPublishGovernanceTransactionMetrics: createMetricsMessage(ncmi, chainID, "out_publish_gov_transaction", func() NodeConnectionMessageMetrics[*iotago.Transaction] {
			return ncmi.GetOutPublishGovernanceTransaction()
		}),
		outPullLatestOutputMetrics: createMetricsMessage(ncmi, chainID, "out_pull_latest_output", func() NodeConnectionMessageMetrics[interface{}] {
			return ncmi.GetOutPullLatestOutput()
		}),
		outPullTxInclusionStateMetrics: createMetricsMessage(ncmi, chainID, "out_pull_tx_inclusion_state", func() NodeConnectionMessageMetrics[iotago.TransactionID] {
			return ncmi.GetOutPullTxInclusionState()
		}),
		outPullOutputByIDMetrics: createMetricsMessage(ncmi, chainID, "out_pull_output_by_id", func() NodeConnectionMessageMetrics[iotago.OutputID] {
			return ncmi.GetOutPullOutputByID()
		}),
		inStateOutputMetrics: createMetricsMessage(ncmi, chainID, "in_state_output", func() NodeConnectionMessageMetrics[*InStateOutput] {
			return ncmi.GetInStateOutput()
		}),
		inAliasOutputMetrics: createMetricsMessage(ncmi, chainID, "in_alias_output", func() NodeConnectionMessageMetrics[*iotago.AliasOutput] {
			return ncmi.GetInAliasOutput()
		}),
		inOutputMetrics: createMetricsMessage(ncmi, chainID, "in_output", func() NodeConnectionMessageMetrics[*InOutput] {
			return ncmi.GetInOutput()
		}),
		inOnLedgerRequestMetrics: createMetricsMessage(ncmi, chainID, "in_on_ledger_request", func() NodeConnectionMessageMetrics[isc.OnLedgerRequest] {
			return ncmi.GetInOnLedgerRequest()
		}),
		inTxInclusionStateMetrics: createMetricsMessage(ncmi, chainID, "in_tx_inclusion_state", func() NodeConnectionMessageMetrics[*TxInclusionStateMsg] {
			return ncmi.GetInTxInclusionState()
		}),
	}
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPublishStateTransaction() NodeConnectionMessageMetrics[*StateTransaction] {
	return ncmmiT.outPublishStateTransactionMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPublishGovernanceTransaction() NodeConnectionMessageMetrics[*iotago.Transaction] {
	return ncmmiT.outPublishGovernanceTransactionMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPullLatestOutput() NodeConnectionMessageMetrics[interface{}] {
	return ncmmiT.outPullLatestOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPullTxInclusionState() NodeConnectionMessageMetrics[iotago.TransactionID] {
	return ncmmiT.outPullTxInclusionStateMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPullOutputByID() NodeConnectionMessageMetrics[iotago.OutputID] {
	return ncmmiT.outPullOutputByIDMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInStateOutput() NodeConnectionMessageMetrics[*InStateOutput] {
	return ncmmiT.inStateOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInAliasOutput() NodeConnectionMessageMetrics[*iotago.AliasOutput] {
	return ncmmiT.inAliasOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInOutput() NodeConnectionMessageMetrics[*InOutput] {
	return ncmmiT.inOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInOnLedgerRequest() NodeConnectionMessageMetrics[isc.OnLedgerRequest] {
	return ncmmiT.inOnLedgerRequestMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInTxInclusionState() NodeConnectionMessageMetrics[*TxInclusionStateMsg] {
	return ncmmiT.inTxInclusionStateMetrics
}
