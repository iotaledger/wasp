// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAckHandler(t *testing.T) {
	t.Parallel()
	n := 10
	nodeIDs := MakeTestNodeIDs("node", n)
	nodesAH := map[NodeID]AckHandler{}
	nodes := map[NodeID]GPA{}
	inputs := map[NodeID]Input{}
	for _, nid := range nodeIDs {
		nodesAH[nid] = NewAckHandler(nid, NewTestRound(nodeIDs, nid), 10*time.Millisecond)
		nodes[nid] = nodesAH[nid]
		inputs[nid] = nil
	}
	tc := NewTestContext(nodes).
		WithInputs(inputs).
		WithInputProbability(0.5).
		WithMessageDeliveryProbability(0.5) // NOTE: The AckHandler has to compensate this.
	tc.RunAll()
	//
	// Tick the timer until all the messages are delivered.
	for {
		allCompleted := true
		for _, n := range nodes {
			if n.Output() == nil {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			for _, n := range nodes {
				require.True(t, *n.Output().(*bool))
			}
			break
		}
		timeTicks := []Message{}
		timestamp := time.Now()
		for _, nid := range nodeIDs {
			timeTicks = append(timeTicks, nodesAH[nid].MakeTickMsg(timestamp))
		}
		tc.WithMessages(timeTicks).RunAll()
	}
}
