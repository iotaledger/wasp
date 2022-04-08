package nodeconnmetrics

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

type nodeConnectionMessagesMetricsImpl struct {
	outPublishTransactionMetrics   NodeConnectionMessageMetrics
	outPullLatestOutputMetrics     NodeConnectionMessageMetrics
	outPullTxInclusionStateMetrics NodeConnectionMessageMetrics
	outPullOutputByIDMetrics       NodeConnectionMessageMetrics

	inOutputMetrics           NodeConnectionMessageMetrics
	inAliasOutputMetrics      NodeConnectionMessageMetrics
	inOnLedgerRequestMetrics  NodeConnectionMessageMetrics
	inTxInclusionStateMetrics NodeConnectionMessageMetrics
}

var _ NodeConnectionMessagesMetrics = &nodeConnectionMessagesMetricsImpl{}

func newNodeConnectionMessagesMetrics(ncmi *nodeConnectionMetricsImpl, chainID *iscp.ChainID) NodeConnectionMessagesMetrics {
	createMessageMetricsFun := func(msgType string, makeRelatedMetricsFun func() NodeConnectionMessageMetrics) NodeConnectionMessageMetrics {
		simpleMessageMetrics := newNodeConnectionMessageSimpleMetrics(ncmi, chainID, msgType)
		if chainID == nil {
			return simpleMessageMetrics
		}
		return newNodeConnectionMessageRelatedMetrics(simpleMessageMetrics, makeRelatedMetricsFun())
	}
	return &nodeConnectionMessagesMetricsImpl{
		outPublishTransactionMetrics:   createMessageMetricsFun("out_publish_transaction", func() NodeConnectionMessageMetrics { return ncmi.GetOutPublishTransaction() }),
		outPullLatestOutputMetrics:     createMessageMetricsFun("out_pull_latest_output", func() NodeConnectionMessageMetrics { return ncmi.GetOutPullLatestOutput() }),
		outPullTxInclusionStateMetrics: createMessageMetricsFun("out_pull_tx_inclusion_state", func() NodeConnectionMessageMetrics { return ncmi.GetOutPullTxInclusionState() }),
		outPullOutputByIDMetrics:       createMessageMetricsFun("out_pull_output_by_id", func() NodeConnectionMessageMetrics { return ncmi.GetOutPullOutputByID() }),

		inOutputMetrics:           createMessageMetricsFun("in_output", func() NodeConnectionMessageMetrics { return ncmi.GetInOutput() }),
		inAliasOutputMetrics:      createMessageMetricsFun("in_alias_output", func() NodeConnectionMessageMetrics { return ncmi.GetInAliasOutput() }),
		inOnLedgerRequestMetrics:  createMessageMetricsFun("in_on_ledger_request", func() NodeConnectionMessageMetrics { return ncmi.GetInOnLedgerRequest() }),
		inTxInclusionStateMetrics: createMessageMetricsFun("in_tx_inclusion_state", func() NodeConnectionMessageMetrics { return ncmi.GetInTxInclusionState() }),
	}
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetOutPublishTransaction() NodeConnectionMessageMetrics {
	return ncmmiT.outPublishTransactionMetrics
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

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInOutput() NodeConnectionMessageMetrics {
	return ncmmiT.inOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInAliasOutput() NodeConnectionMessageMetrics {
	return ncmmiT.inAliasOutputMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInOnLedgerRequest() NodeConnectionMessageMetrics {
	return ncmmiT.inOnLedgerRequestMetrics
}

func (ncmmiT *nodeConnectionMessagesMetricsImpl) GetInTxInclusionState() NodeConnectionMessageMetrics {
	return ncmmiT.inTxInclusionStateMetrics
}
