package model

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type NodeConnectionMessageMetrics struct {
	Total       uint32    `json:"total" swagger:"desc(Total number of messages sent/received)"`
	LastEvent   time.Time `json:"lastEvent" swagger:"desc(Last time the message was sent/received)"`
	LastMessage string    `json:"lastMessage" swagger:"desc(The print out of the last message)"`
}

type NodeConnectionMessagesMetrics struct {
	OutPublishStateTransaction      *NodeConnectionMessageMetrics `json:"outPublishStateTransaction" swagger:"desc(Stats of sent out PublishStateTransaction messages)"`
	OutPublishGovernanceTransaction *NodeConnectionMessageMetrics `json:"outPublishGovernanceTransaction" swagger:"desc(Stats of sent out PublishGovernanceTransaction messages)"`
	OutPullLatestOutput             *NodeConnectionMessageMetrics `json:"outPullLatestOutput" swagger:"desc(Stats of sent out PullLatestOutput messages)"`
	OutPullTxInclusionState         *NodeConnectionMessageMetrics `json:"outPullTxInclusionState" swagger:"desc(Stats of sent out PullTxInclusionState messages)"`
	OutPullOutputByID               *NodeConnectionMessageMetrics `json:"outPullOutputByID" swagger:"desc(Stats of sent out PullOutputByID messages)"`
	InStateOutput                   *NodeConnectionMessageMetrics `json:"inStateOutput" swagger:"desc(Stats of received State output messages)"`
	InAliasOutput                   *NodeConnectionMessageMetrics `json:"inAliasOutput" swagger:"desc(Stats of received AliasOutput messages)"`
	InOutput                        *NodeConnectionMessageMetrics `json:"inOutput" swagger:"desc(Stats of received Output messages)"`
	InOnLedgerRequest               *NodeConnectionMessageMetrics `json:"inOnLedgerRequest" swagger:"desc(Stats of received OnLedgerRequest messages)"`
	InTxInclusionState              *NodeConnectionMessageMetrics `json:"inTxInclusionState" swagger:"desc(Stats of received TxInclusionState messages)"`
}

type NodeConnectionMetrics struct {
	NodeConnectionMessagesMetrics `json:"nodeConnectionMessagesMetrics"`
	InMilestone                   *NodeConnectionMessageMetrics `json:"inMilestone" swagger:"desc(Stats of received Milestone messages)"`
	Registered                    []ChainIDBech32               `json:"registered" swagger:"desc(Chain IDs of the chains registered to receiving L1 events)"`
}

func NewNodeConnectionMetrics(metrics nodeconnmetrics.NodeConnectionMetrics) *NodeConnectionMetrics {
	ncmm := NewNodeConnectionMessagesMetrics(metrics)
	registered := metrics.GetRegistered()
	r := make([]ChainIDBech32, len(registered))
	for i := range r {
		r[i] = NewChainIDBech32(registered[i])
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
