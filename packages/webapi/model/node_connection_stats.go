package model

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
)

type NodeConnectionMessageStats struct {
	Total       uint32    `swagger:"desc(Total number of messages sent/received)"`
	LastEvent   time.Time `swagger:"desc(Last time the message was sent/received)"`
	LastMessage string    `swagger:"desc(The print out of the last message)"`
}

type NodeConnectionMessagesStats struct {
	OutPullState                     *NodeConnectionMessageStats `swagger:"desc(Stats of sent out PullState messages)"`
	OutPullTransactionInclusionState *NodeConnectionMessageStats `swagger:"desc(Stats of sent out PullTransactionInclusionState messages)"`
	OutPullConfirmedOutput           *NodeConnectionMessageStats `swagger:"desc(Stats of sent out PullConfirmedOutput messages)"`
	OutPostTransaction               *NodeConnectionMessageStats `swagger:"desc(Stats of sent out PostTransaction messages)"`

	InTransaction        *NodeConnectionMessageStats `swagger:"desc(Stats of received Transaction messages)"`
	InInclusionState     *NodeConnectionMessageStats `swagger:"desc(Stats of received InclusionState messages)"`
	InOutput             *NodeConnectionMessageStats `swagger:"desc(Stats of received Output messages)"`
	InUnspentAliasOutput *NodeConnectionMessageStats `swagger:"desc(Stats of received UnspentAliasOutput messages)"`
}

type NodeConnectionStats struct {
	NodeConnectionMessagesStats
	Subscribed []Address
}

func NewNodeConnectionStats(stats *chain.NodeConnectionStats) *NodeConnectionStats {
	ncms := NewNodeConnectionMessagesStats(&stats.NodeConnectionMessagesStats)
	s := make([]Address, len(stats.Subscribed))
	for i := range s {
		s[i] = NewAddress(stats.Subscribed[i])
	}
	return &NodeConnectionStats{
		NodeConnectionMessagesStats: *ncms,
		Subscribed:                  s,
	}
}

func NewNodeConnectionMessagesStats(stats *chain.NodeConnectionMessagesStats) *NodeConnectionMessagesStats {
	return &NodeConnectionMessagesStats{
		OutPullState:                     NewNodeConnectionMessageStats(&stats.OutPullState),
		OutPullTransactionInclusionState: NewNodeConnectionMessageStats(&stats.OutPullTransactionInclusionState),
		OutPullConfirmedOutput:           NewNodeConnectionMessageStats(&stats.OutPullConfirmedOutput),
		OutPostTransaction:               NewNodeConnectionMessageStats(&stats.OutPostTransaction),

		InTransaction:        NewNodeConnectionMessageStats(&stats.InTransaction),
		InInclusionState:     NewNodeConnectionMessageStats(&stats.InInclusionState),
		InOutput:             NewNodeConnectionMessageStats(&stats.InOutput),
		InUnspentAliasOutput: NewNodeConnectionMessageStats(&stats.InUnspentAliasOutput),
	}
}

func NewNodeConnectionMessageStats(stats *chain.NodeConnectionMessageStats) *NodeConnectionMessageStats {
	return &NodeConnectionMessageStats{
		Total:       stats.Total.Load(),
		LastEvent:   stats.LastEvent,
		LastMessage: fmt.Sprintf("%v", stats.LastMessage),
	}
}
