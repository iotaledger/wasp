package nodeconnmetrics

type emptyNodeConnectionMessagesMetrics struct {
	empty NodeConnectionMessageMetrics
}

var _ NodeConnectionMessagesMetrics = &emptyNodeConnectionMessagesMetrics{}

func NewEmptyNodeConnectionMessagesMetrics() NodeConnectionMessagesMetrics {
	return &emptyNodeConnectionMessagesMetrics{empty: newEmptyNodeConnectionMessageMetrics()}
}

func (ecmm *emptyNodeConnectionMessagesMetrics) GetOutPullState() NodeConnectionMessageMetrics {
	return ecmm.empty
}

func (ecmm *emptyNodeConnectionMessagesMetrics) GetOutPullTransactionInclusionState() NodeConnectionMessageMetrics {
	return ecmm.empty
}

func (ecmm *emptyNodeConnectionMessagesMetrics) GetOutPullConfirmedOutput() NodeConnectionMessageMetrics {
	return ecmm.empty
}

func (ecmm *emptyNodeConnectionMessagesMetrics) GetOutPostTransaction() NodeConnectionMessageMetrics {
	return ecmm.empty
}

func (ecmm *emptyNodeConnectionMessagesMetrics) GetInTransaction() NodeConnectionMessageMetrics {
	return ecmm.empty
}

func (ecmm *emptyNodeConnectionMessagesMetrics) GetInInclusionState() NodeConnectionMessageMetrics {
	return ecmm.empty
}

func (ecmm *emptyNodeConnectionMessagesMetrics) GetInOutput() NodeConnectionMessageMetrics {
	return ecmm.empty
}

func (ecmm *emptyNodeConnectionMessagesMetrics) GetInUnspentAliasOutput() NodeConnectionMessageMetrics {
	return ecmm.empty
}
