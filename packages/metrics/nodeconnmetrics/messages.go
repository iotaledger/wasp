package nodeconnmetrics

import (
	"github.com/iotaledger/wasp/packages/isc"
)

type nodeConnectionMessagesMetricsImpl struct {
	outPublishStateTransactionMetrics      NodeConnectionMessageMetrics
	outPublishGovernanceTransactionMetrics NodeConnectionMessageMetrics
	outPullLatestOutputMetrics             NodeConnectionMessageMetrics
	outPullTxInclusionStateMetrics         NodeConnectionMessageMetrics
	outPullOutputByIDMetrics               NodeConnectionMessageMetrics

	inStateOutputMetrics      NodeConnectionMessageMetrics
	inAliasOutputMetrics      NodeConnectionMessageMetrics
	inOutputMetrics           NodeConnectionMessageMetrics
	inOnLedgerRequestMetrics  NodeConnectionMessageMetrics
	inTxInclusionStateMetrics NodeConnectionMessageMetrics
}

var _ NodeConnectionMessagesMetrics = &nodeConnectionMessagesMetricsImpl{}

func newNodeConnectionMessagesMetrics(ncmi *nodeConnectionMetricsImpl, chainID *isc.ChainID) NodeConnectionMessagesMetrics {
	createMessageMetricsFun := func(msgType string, makeRelatedMetricsFun func() NodeConnectionMessageMetrics) NodeConnectionMessageMetrics {
		simpleMessageMetrics := newNodeConnectionMessageSimpleMetrics(ncmi, chainID, msgType)
		if chainID == nil {
			return simpleMessageMetrics
		}
		return newNodeConnectionMessageRelatedMetrics(simpleMessageMetrics, makeRelatedMetricsFun())
	}
	return &nodeConnectionMessagesMetricsImpl{
		outPublishStateTransactionMetrics:      createMessageMetricsFun("out_publish_state_transaction", func() NodeConnectionMessageMetrics { return ncmi.GetOutPublishStateTransaction() }),
		outPublishGovernanceTransactionMetrics: createMessageMetricsFun("out_publish_gov_transaction", func() NodeConnectionMessageMetrics { return ncmi.GetOutPublishGovernanceTransaction() }),
		outPullLatestOutputMetrics:             createMessageMetricsFun("out_pull_latest_output", func() NodeConnectionMessageMetrics { return ncmi.GetOutPullLatestOutput() }),
		outPullTxInclusionStateMetrics:         createMessageMetricsFun("out_pull_tx_inclusion_state", func() NodeConnectionMessageMetrics { return ncmi.GetOutPullTxInclusionState() }),
		outPullOutputByIDMetrics:               createMessageMetricsFun("out_pull_output_by_id", func() NodeConnectionMessageMetrics { return ncmi.GetOutPullOutputByID() }),

		inStateOutputMetrics:      createMessageMetricsFun("in_state_output", func() NodeConnectionMessageMetrics { return ncmi.GetInStateOutput() }),
		inAliasOutputMetrics:      createMessageMetricsFun("in_alias_output", func() NodeConnectionMessageMetrics { return ncmi.GetInAliasOutput() }),
		inOutputMetrics:           createMessageMetricsFun("in_output", func() NodeConnectionMessageMetrics { return ncmi.GetInOutput() }),
		inOnLedgerRequestMetrics:  createMessageMetricsFun("in_on_ledger_request", func() NodeConnectionMessageMetrics { return ncmi.GetInOnLedgerRequest() }),
		inTxInclusionStateMetrics: createMessageMetricsFun("in_tx_inclusion_state", func() NodeConnectionMessageMetrics { return ncmi.GetInTxInclusionState() }),
	}
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPublishStateTransaction() NodeConnectionMessageMetrics {
	return ncmmiT.outPublishStateTransactionMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPublishGovernanceTransaction() NodeConnectionMessageMetrics {
	return ncmmiT.outPublishGovernanceTransactionMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPullLatestOutput() NodeConnectionMessageMetrics {
	return ncmmiT.outPullLatestOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPullTxInclusionState() NodeConnectionMessageMetrics {
	return ncmmiT.outPullTxInclusionStateMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPullOutputByID() NodeConnectionMessageMetrics {
	return ncmmiT.outPullOutputByIDMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInStateOutput() NodeConnectionMessageMetrics {
	return ncmmiT.inStateOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInAliasOutput() NodeConnectionMessageMetrics {
	return ncmmiT.inAliasOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInOutput() NodeConnectionMessageMetrics {
	return ncmmiT.inOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInOnLedgerRequest() NodeConnectionMessageMetrics {
	return ncmmiT.inOnLedgerRequestMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInTxInclusionState() NodeConnectionMessageMetrics {
	return ncmmiT.inTxInclusionStateMetrics
}
