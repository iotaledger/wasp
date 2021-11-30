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
	OutPullState                     *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullState messages)"`
	OutPullTransactionInclusionState *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullTransactionInclusionState messages)"`
	OutPullConfirmedOutput           *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PullConfirmedOutput messages)"`
	OutPostTransaction               *NodeConnectionMessageMetrics `swagger:"desc(Stats of sent out PostTransaction messages)"`

	InTransaction        *NodeConnectionMessageMetrics `swagger:"desc(Stats of received Transaction messages)"`
	InInclusionState     *NodeConnectionMessageMetrics `swagger:"desc(Stats of received InclusionState messages)"`
	InOutput             *NodeConnectionMessageMetrics `swagger:"desc(Stats of received Output messages)"`
	InUnspentAliasOutput *NodeConnectionMessageMetrics `swagger:"desc(Stats of received UnspentAliasOutput messages)"`
}

type NodeConnectionMetrics struct {
	NodeConnectionMessagesMetrics
	Subscribed []Address
}

func NewNodeConnectionMetrics(metrics nodeconnmetrics.NodeConnectionMetrics) *NodeConnectionMetrics {
	ncmm := NewNodeConnectionMessagesMetrics(metrics)
	subscribed := metrics.GetSubscribed()
	s := make([]Address, len(subscribed))
	for i := range s {
		s[i] = NewAddress(subscribed[i])
	}
	return &NodeConnectionMetrics{
		NodeConnectionMessagesMetrics: *ncmm,
		Subscribed:                    s,
	}
}

func NewNodeConnectionMessagesMetrics(metrics nodeconnmetrics.NodeConnectionMessagesMetrics) *NodeConnectionMessagesMetrics {
	return &NodeConnectionMessagesMetrics{
		OutPullState:                     NewNodeConnectionMessageMetrics(metrics.GetOutPullState()),
		OutPullTransactionInclusionState: NewNodeConnectionMessageMetrics(metrics.GetOutPullTransactionInclusionState()),
		OutPullConfirmedOutput:           NewNodeConnectionMessageMetrics(metrics.GetOutPullConfirmedOutput()),
		OutPostTransaction:               NewNodeConnectionMessageMetrics(metrics.GetOutPostTransaction()),

		InTransaction:        NewNodeConnectionMessageMetrics(metrics.GetInTransaction()),
		InInclusionState:     NewNodeConnectionMessageMetrics(metrics.GetInInclusionState()),
		InOutput:             NewNodeConnectionMessageMetrics(metrics.GetInOutput()),
		InUnspentAliasOutput: NewNodeConnectionMessageMetrics(metrics.GetInUnspentAliasOutput()),
	}
}

func NewNodeConnectionMessageMetrics(metrics nodeconnmetrics.NodeConnectionMessageMetrics) *NodeConnectionMessageMetrics {
	return &NodeConnectionMessageMetrics{
		Total:       metrics.GetMessageTotal(),
		LastEvent:   metrics.GetLastEvent(),
		LastMessage: fmt.Sprintf("%v", metrics.GetLastMessage()),
	}
}
