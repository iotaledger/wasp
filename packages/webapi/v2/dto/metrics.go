package dto

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain/nodeconnchain"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type MetricItem struct {
	Timestamp time.Time
	Total     uint32
}

func MapMetricItem(metrics nodeconnmetrics.NodeConnectionMessageMetrics) MetricItem {
	return MetricItem{
		Timestamp: metrics.GetLastEvent(),
		Total:     metrics.GetMessageTotal(),
	}
}

type PullOutputByID struct {
	MetricItem
	OutputID *iotago.UTXOInput
}

func MapPullOutputByID(metrics nodeconnmetrics.NodeConnectionMessageMetrics) *PullOutputByID {
	outputID, _ := metrics.GetLastMessage().(*iotago.UTXOInput)

	return &PullOutputByID{
		MetricItem: MapMetricItem(metrics),
		OutputID:   outputID,
	}
}

type PullTXInclusionStateOut struct {
	MetricItem
	TransactionID iotago.TransactionID
}

func MapPullTXInclusionStateOut(metrics nodeconnmetrics.NodeConnectionMessageMetrics) *PullTXInclusionStateOut {
	transactionID, _ := metrics.GetLastMessage().(iotago.TransactionID)

	return &PullTXInclusionStateOut{
		MetricItem:    MapMetricItem(metrics),
		TransactionID: transactionID,
	}
}

type PullTXInclusionStateIn struct {
	MetricItem
	State *nodeconnchain.TxInclusionStateMsg
}

func MapPullTXInclusionStateIn(metrics nodeconnmetrics.NodeConnectionMessageMetrics) *PullTXInclusionStateIn {
	state, _ := metrics.GetLastMessage().(*nodeconnchain.TxInclusionStateMsg)

	return &PullTXInclusionStateIn{
		MetricItem: MapMetricItem(metrics),
		State:      state,
	}
}

type PublishStateTransaction struct {
	MetricItem
	State nodeconnchain.StateTransaction
}

func MapPublishStateTransaction(metrics nodeconnmetrics.NodeConnectionMessageMetrics) *PublishStateTransaction {
	state, _ := metrics.GetLastMessage().(nodeconnchain.StateTransaction)

	return &PublishStateTransaction{
		MetricItem: MapMetricItem(metrics),
		State:      state,
	}
}

type NodeConnectionMetrics struct {
	PullOutputByID          *PullOutputByID
	PullTXInclusionStateOut *PullTXInclusionStateOut
	PullTXInclusionStateIn  *PullTXInclusionStateIn
	PullLatestOutput        *MetricItem
	PublishStateTransaction *PublishStateTransaction
}

func k() {
	var metrics nodeconnmetrics.NodeConnectionMetrics
	// metrics.GetOutPullOutputByID()
	// metrics.GetOutPullTxInclusionState()
	// metrics.GetOutPullLatestOutput()
	// metrics.GetInTxInclusionState()
	metrics.GetOutPublishStateTransaction()
}
