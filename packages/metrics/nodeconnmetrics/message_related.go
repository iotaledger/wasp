package nodeconnmetrics

type nodeConnectionMessageRelatedMetricsImpl struct {
	NodeConnectionMessageMetrics
	related NodeConnectionMessageMetrics
}

var _ NodeConnectionMessageMetrics = &nodeConnectionMessageRelatedMetricsImpl{}

func newNodeConnectionMessageRelatedMetrics(metrics, related NodeConnectionMessageMetrics) NodeConnectionMessageMetrics {
	return &nodeConnectionMessageRelatedMetricsImpl{
		NodeConnectionMessageMetrics: metrics,
		related:                      related,
	}
}

func (ncmrmi *nodeConnectionMessageRelatedMetricsImpl) CountLastMessage(msg interface{}) {
	ncmrmi.NodeConnectionMessageMetrics.CountLastMessage(msg)
	ncmrmi.related.CountLastMessage(msg)
}
