package dto

import (
	"time"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
)

type MetricItem[T interface{}] struct {
	Messages    uint32
	Timestamp   time.Time
	LastMessage T
}

type ChainMessageMetrics struct {
	InAnchor          *MetricItem[*metrics.StateAnchor]
	InOnLedgerRequest *MetricItem[isc.OnLedgerRequest]

	OutPublishStateTransaction *MetricItem[*metrics.StateTransaction]
}

type NodeMessageMetrics struct {
	RegisteredChainIDs []isc.ChainID

	InAnchor          *MetricItem[*metrics.StateAnchor]
	InOnLedgerRequest *MetricItem[isc.OnLedgerRequest]

	OutPublishStateTransaction *MetricItem[*metrics.StateTransaction]
}

func MapMetricItem[T interface{}](metrics metrics.IMessageMetric[T]) *MetricItem[T] {
	return &MetricItem[T]{
		Messages:    metrics.MessagesTotal(),
		Timestamp:   metrics.LastMessageTime(),
		LastMessage: metrics.LastMessage(),
	}
}
