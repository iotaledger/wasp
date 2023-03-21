package dto

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
)

type MetricItem[T interface{}] struct {
	Messages    uint32
	Timestamp   time.Time
	LastMessage T
}

type ChainMessageMetrics struct {
	InStateOutput      *MetricItem[*metrics.InStateOutput]
	InAliasOutput      *MetricItem[*iotago.AliasOutput]
	InOutput           *MetricItem[*metrics.InOutput]
	InOnLedgerRequest  *MetricItem[isc.OnLedgerRequest]
	InTxInclusionState *MetricItem[*metrics.TxInclusionStateMsg]

	OutPublishStateTransaction      *MetricItem[*metrics.StateTransaction]
	OutPublishGovernanceTransaction *MetricItem[*iotago.Transaction]
	OutPullLatestOutput             *MetricItem[interface{}]
	OutPullTxInclusionState         *MetricItem[iotago.TransactionID]
	OutPullOutputByID               *MetricItem[iotago.OutputID]
}

type NodeMessageMetrics struct {
	RegisteredChainIDs []isc.ChainID

	InMilestone        *MetricItem[*nodeclient.MilestoneInfo]
	InStateOutput      *MetricItem[*metrics.InStateOutput]
	InAliasOutput      *MetricItem[*iotago.AliasOutput]
	InOutput           *MetricItem[*metrics.InOutput]
	InOnLedgerRequest  *MetricItem[isc.OnLedgerRequest]
	InTxInclusionState *MetricItem[*metrics.TxInclusionStateMsg]

	OutPublishStateTransaction      *MetricItem[*metrics.StateTransaction]
	OutPublishGovernanceTransaction *MetricItem[*iotago.Transaction]
	OutPullLatestOutput             *MetricItem[interface{}]
	OutPullTxInclusionState         *MetricItem[iotago.TransactionID]
	OutPullOutputByID               *MetricItem[iotago.OutputID]
}

func MapMetricItem[T interface{}](metrics metrics.IMessageMetric[T]) *MetricItem[T] {
	return &MetricItem[T]{
		Messages:    metrics.MessagesTotal(),
		Timestamp:   metrics.LastMessageTime(),
		LastMessage: metrics.LastMessage(),
	}
}
