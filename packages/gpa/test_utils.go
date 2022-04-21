// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"
	"math/rand"
)

// RunWithInputs is to be used in test cases.
func RunTestWithInputs(nodes map[NodeID]GPA, inputs map[NodeID]Input) {
	msgs := []Message{}
	nodeIDs := []NodeID{}
	for nid := range inputs {
		nodeIDs = append(nodeIDs, nid)
	}
	nodeIDs = ShuffleNodeIDs(nodeIDs)
	for _, nid := range nodeIDs {
		msgs = append(msgs, nodes[nid].Input(inputs[nid])...)
	}
	RunTestWithMessages(nodes, msgs)
}

// RunWithMessages is to be used in test cases.
func RunTestWithMessages(nodes map[NodeID]GPA, msgs []Message) {
	for len(msgs) > 0 {
		msgIdx := rand.Intn(len(msgs))
		msg := msgs[msgIdx]
		msgs = append(msgs[:msgIdx], msgs[msgIdx+1:]...)
		msgs = append(msgs, nodes[msg.Recipient()].Message(msg)...)
	}
}

func MakeTestNodeIDs(prefix string, n int) []NodeID {
	nodeIDs := make([]NodeID, n)
	for i := range nodeIDs {
		nodeIDs[i] = NodeID(fmt.Sprintf("%s-%03d", prefix, i))
	}
	return nodeIDs
}

func ShuffleNodeIDs(nodeIDs []NodeID) []NodeID {
	rand.Shuffle(len(nodeIDs), func(i, j int) { nodeIDs[i], nodeIDs[j] = nodeIDs[j], nodeIDs[i] })
	return nodeIDs
}

func CopyNodeIDs(nodeIDs []NodeID) []NodeID {
	c := make([]NodeID, len(nodeIDs))
	for i := range c {
		c[i] = nodeIDs[i]
	}
	return c
}
