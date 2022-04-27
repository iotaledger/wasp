// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"
	"math/rand"
)

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
