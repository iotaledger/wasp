package model

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type NodeConnectionMessageMetrics struct {
	Total       uint32    `swagger:"desc(Total number of messages sent/received)"`
	LastEvent   time.Time `swagger:"desc(Last time the message was sent/received)"`
	LastMessage string    `swagger:"desc(The print out of the last message)"`
}

type NodeConnectionMessagesMetrics struct {
	OutPublishStateTransaction      *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PublishStateTransaction messages)"`
	OutPublishGovernanceTransaction *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PublishGovernanceTransaction messages)"`
	OutPullLatestOutput             *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullLatestOutput messages)"`
	OutPullTxInclusionState         *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullTxInclusionState messages)"`
	OutPullOutputByID               *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullOutputByID messages)"`

	InStateOutput      *NodeConnectionMessageMetrics `swagger:"desc(Stats of received State output messages)"`
	InAliasOutput      *NodeConnectionMessageMetrics `swagger:"desc(Stats of received AliasOutput messages)"`
	InOutput           *NodeConnectionMessageMetrics `swagger:"desc(Stats of received Output messages)"`
	InOnLedgerRequest  *NodeConnectionMessageMetrics `swagger:"desc(Stats of received OnLedgerRequest messages)"`
	InTxInclusionState *NodeConnectionMessageMetrics `swagger:"desc(Stats of received TxInclusionState messages)"`
}

type NodeConnectionMetrics struct {
	NodeConnectionMessagesMetrics
	InMilestone *NodeConnectionMessageMetrics `swagger:"desc(Stats of received Milestone messages)"`
	Registered  []ChainID                     `swagger:"desc(Chain IDs of the chains registered to receiving L1 events)"`
}

func NewNodeConnectionMetrics(metrics nodeconnmetrics.NodeConnectionMetrics) *NodeConnectionMetrics {
	ncmm := NewNodeConnectionMessagesMetrics(metrics)
	registered := metrics.GetRegistered()
	r := make([]ChainID, len(registered))
	for i := range r {
		r[i] = NewChainID(registered[i])
	}
	return &NodeConnectionMetrics{
		NodeConnectionMessagesMetrics: *ncmm,
		InMilestone:                   NewNodeConnectionMessageMetrics(metrics.GetInMilestone()),
		Registered:                    r,
	}
}

func NewNodeConnectionMessagesMetrics(metrics nodeconnmetrics.NodeConnectionMessagesMetrics) *NodeConnectionMessagesMetrics {
	return &NodeConnectionMessagesMetrics{
		OutPublishStateTransaction:      NewNodeConnectionMessageMetrics(metrics.GetOutPublishStateTransaction()),
		OutPublishGovernanceTransaction: NewNodeConnectionMessageMetrics(metrics.GetOutPublishGovernanceTransaction()),
		OutPullLatestOutput:             NewNodeConnectionMessageMetrics(metrics.GetOutPullLatestOutput()),
		OutPullTxInclusionState:         NewNodeConnectionMessageMetrics(metrics.GetOutPullTxInclusionState()),
		OutPullOutputByID:               NewNodeConnectionMessageMetrics(metrics.GetOutPullOutputByID()),

		InStateOutput:      NewNodeConnectionMessageMetrics(metrics.GetInStateOutput()),
		InAliasOutput:      NewNodeConnectionMessageMetrics(metrics.GetInAliasOutput()),
		InOutput:           NewNodeConnectionMessageMetrics(metrics.GetInOutput()),
		InOnLedgerRequest:  NewNodeConnectionMessageMetrics(metrics.GetInOnLedgerRequest()),
		InTxInclusionState: NewNodeConnectionMessageMetrics(metrics.GetInTxInclusionState()),
	}
}

func NewNodeConnectionMessageMetrics[T any](metrics nodeconnmetrics.NodeConnectionMessageMetrics[T]) *NodeConnectionMessageMetrics {
	return &NodeConnectionMessageMetrics{
		Total:       metrics.GetMessageTotal(),
		LastEvent:   metrics.GetLastEvent(),
		LastMessage: fmt.Sprintf("%v", metrics.GetLastMessage()),
	}
}
