package nodeconnmetrics

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

type nodeConnectionMessagesMetricsImpl struct {
	outPullStateMetrics                     NodeConnectionMessageMetrics
	outPullTransactionInclusionStateMetrics NodeConnectionMessageMetrics
	outPullConfirmedOutputMetrics           NodeConnectionMessageMetrics
	outPostTransactionMetrics               NodeConnectionMessageMetrics

	inTransactionMetrics        NodeConnectionMessageMetrics
	inInclusionStateMetrics     NodeConnectionMessageMetrics
	inOutputMetrics             NodeConnectionMessageMetrics
	inUnspentAliasOutputMetrics NodeConnectionMessageMetrics
}

var _ NodeConnectionMessagesMetrics = &nodeConnectionMessagesMetricsImpl{}

func newNodeConnectionMessagesMetrics(ncmi *nodeConnectionMetricsImpl, chainID *iscp.ChainID) NodeConnectionMessagesMetrics {
	return &nodeConnectionMessagesMetricsImpl{
		outPullStateMetrics:                     newNodeConnectionMessageMetrics(ncmi, chainID, "out_pull_state"),
		outPullTransactionInclusionStateMetrics: newNodeConnectionMessageMetrics(ncmi, chainID, "out_pull_transaction_inclusion_state"),
		outPullConfirmedOutputMetrics:           newNodeConnectionMessageMetrics(ncmi, chainID, "out_pull_confirmed_output"),
		outPostTransactionMetrics:               newNodeConnectionMessageMetrics(ncmi, chainID, "out_post_transaction"),

		inTransactionMetrics:        newNodeConnectionMessageMetrics(ncmi, chainID, "in_transaction"),
		inInclusionStateMetrics:     newNodeConnectionMessageMetrics(ncmi, chainID, "in_inclusion_state"),
		inOutputMetrics:             newNodeConnectionMessageMetrics(ncmi, chainID, "in_output"),
		inUnspentAliasOutputMetrics: newNodeConnectionMessageMetrics(ncmi, chainID, "in_alias_output"),
	}
}

func (ncmmi *nodeConnectionMessagesMetricsImpl) GetOutPullState() NodeConnectionMessageMetrics {
	return ncmmi.outPullStateMetrics
}

func (ncmmi *nodeConnectionMessagesMetricsImpl) GetOutPullTransactionInclusionState() NodeConnectionMessageMetrics {
	return ncmmi.outPullTransactionInclusionStateMetrics
}

func (ncmmi *nodeConnectionMessagesMetricsImpl) GetOutPullConfirmedOutput() NodeConnectionMessageMetrics {
	return ncmmi.outPullConfirmedOutputMetrics
}

func (ncmmi *nodeConnectionMessagesMetricsImpl) GetOutPostTransaction() NodeConnectionMessageMetrics {
	return ncmmi.outPostTransactionMetrics
}

func (ncmmi *nodeConnectionMessagesMetricsImpl) GetInTransaction() NodeConnectionMessageMetrics {
	return ncmmi.inTransactionMetrics
}

func (ncmmi *nodeConnectionMessagesMetricsImpl) GetInInclusionState() NodeConnectionMessageMetrics {
	return ncmmi.inInclusionStateMetrics
}

func (ncmmi *nodeConnectionMessagesMetricsImpl) GetInOutput() NodeConnectionMessageMetrics {
	return ncmmi.inOutputMetrics
}

func (ncmmi *nodeConnectionMessagesMetricsImpl) GetInUnspentAliasOutput() NodeConnectionMessageMetrics {
	return ncmmi.inUnspentAliasOutputMetrics
}
