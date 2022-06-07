package nodeconnmetrics

type emptyNodeConnectionMessagesMetrics struct {
	emptyMessageMetrics NodeConnectionMessageMetrics
}

var _ NodeConnectionMessagesMetrics = &emptyNodeConnectionMessagesMetrics{}

func newEmptyNodeConnectionMessagesMetrics() *emptyNodeConnectionMessagesMetrics {
	return &emptyNodeConnectionMessagesMetrics{emptyMessageMetrics: newEmptyNodeConnectionMessageMetrics()}
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPublishTransaction() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPullLatestOutput() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPullTxInclusionState() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetOutPullOutputByID() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInStateOutput() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInAliasOutput() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInOutput() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInOnLedgerRequest() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}

func (encmmT *emptyNodeConnectionMessagesMetrics) GetInTxInclusionState() NodeConnectionMessageMetrics {
	return encmmT.emptyMessageMetrics
}
