package models

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type MetricItem[T any] struct {
	Messages    uint32
	Timestamp   time.Time
	LastMessage T
}

type ChainMetrics struct {
	InAliasOutput                   *MetricItem[*iotago.AliasOutput]
	InOnLedgerRequest               *MetricItem[isc.OnLedgerRequest]
	InOutput                        *MetricItem[*nodeconnmetrics.InOutput]
	InStateOutput                   *MetricItem[*nodeconnmetrics.InStateOutput]
	InTxInclusionState              *MetricItem[*nodeconnmetrics.TxInclusionStateMsg]
	OutPublishGovernanceTransaction *MetricItem[*iotago.Transaction]
	OutPullLatestOutput             *MetricItem[interface{}]
	OutPullOutputByID               *MetricItem[*iotago.UTXOInput]
	OutPullTxInclusionState         *MetricItem[iotago.TransactionID]
}
