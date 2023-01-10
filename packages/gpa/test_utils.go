// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"encoding/binary"
	"math/rand"
)

func MakeTestNodeIDFromIndex(index int) NodeID {
	nodeID := NodeID{}

	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, uint32(index))
	copy(nodeID[:4], indexBytes)

	return nodeID
}

func MakeTestNodeIDs(n int) []NodeID {
	nodeIDs := make([]NodeID, n)
	for i := range nodeIDs {
		nodeIDs[i] = MakeTestNodeIDFromIndex(i)
	}
	return nodeIDs
}

func RandomTestNodeID() NodeID {
	return MakeTestNodeIDFromIndex(rand.Int())
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
