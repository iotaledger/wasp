package model

import (
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type NodeConnectionMessageMetrics struct {
	Total       uint32    `swagger:"desc(Total number of messages sent/received)"`
	LastEvent   time.Time `swagger:"desc(Last time the message was sent/received)"`
	LastMessage string    `swagger:"desc(The print out of the last message)"`
}

type NodeConnectionMessagesMetrics struct {
	OutPublishTransaction   *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PublishTransaction messages)"`
	OutPullLatestOutput     *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullLatestOutput messages)"`
	OutPullTxInclusionState *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullTxInclusionState messages)"`
	OutPullOutputByID       *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullOutputByID messages)"`

	InOutput             *NodeConnectionMessageMetrics `swagger:"desc(Stats of received Output messages)"`
	InAliasOutput      *NodeConnectionMessageMetrics `swagger:"desc(Stats of received AliasOutput messages)"`
	InOnLedgerRequest  *NodeConnectionMessageMetrics `swagger:"desc(Stats of received OnLedgerRequest messages)"`
	InTxInclusionState *NodeConnectionMessageMetrics `swagger:"desc(Stats of received TxInclusionState messages)"`
}

type NodeConnectionMetrics struct {
	NodeConnectionMessagesMetrics
	InMilestone *NodeConnectionMessageMetrics `swagger:"desc(Stats of received Milestone messages)"`
	Registered  []Address                     `swagger:"desc(Addresses of the chains registered to receiving L1 events)"`
}

func NewNodeConnectionMetrics(metrics nodeconnmetrics.NodeConnectionMetrics, networkPrefix iotago.NetworkPrefix) *NodeConnectionMetrics {
	ncmm := NewNodeConnectionMessagesMetrics(metrics)
	registered := metrics.GetRegistered()
	r := make([]Address, len(registered))
	for i := range r {
		r[i] = NewAddress(registered[i], networkPrefix)
	}
	return &NodeConnectionMetrics{
		NodeConnectionMessagesMetrics: *ncmm,
		InMilestone:                   NewNodeConnectionMessageMetrics(metrics.GetInMilestone()),
		Registered:                    r,
	}
}

func NewNodeConnectionMessagesMetrics(metrics nodeconnmetrics.NodeConnectionMessagesMetrics) *NodeConnectionMessagesMetrics {
	return &NodeConnectionMessagesMetrics{
		OutPublishTransaction:   NewNodeConnectionMessageMetrics(metrics.GetOutPublishTransaction()),
		OutPullLatestOutput:     NewNodeConnectionMessageMetrics(metrics.GetOutPullLatestOutput()),
		OutPullTxInclusionState: NewNodeConnectionMessageMetrics(metrics.GetOutPullTxInclusionState()),
		OutPullOutputByID:       NewNodeConnectionMessageMetrics(metrics.GetOutPullOutputByID()),

		InOutput:             NewNodeConnectionMessageMetrics(metrics.GetInOutput()),
		InAliasOutput:      NewNodeConnectionMessageMetrics(metrics.GetInAliasOutput()),
		InOnLedgerRequest:  NewNodeConnectionMessageMetrics(metrics.GetInOnLedgerRequest()),
		InTxInclusionState: NewNodeConnectionMessageMetrics(metrics.GetInTxInclusionState()),
	}
}

func NewNodeConnectionMessageMetrics(metrics nodeconnmetrics.NodeConnectionMessageMetrics) *NodeConnectionMessageMetrics {
	return &NodeConnectionMessageMetrics{
		Total:       metrics.GetMessageTotal(),
		LastEvent:   metrics.GetLastEvent(),
		LastMessage: fmt.Sprintf("%v", metrics.GetLastMessage()),
	}
}
