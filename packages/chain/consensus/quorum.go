// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// In this file we have misc functions for calculating quorum sizes.

package consensus

// MaxNodesFaulty calculates maximal number of nodes that
// can be assumed faulty F given a total number of nodes N.
func MaxNodesFaulty(n int) int {
	return (n - 1) / 3
}

// MinNodesInQuorum returns a minimal number T of nodes required for a quorum.
// If N=3F+1 then T=2F+1, but if if N is arbitrary, then T=N-F.
func MinNodesInQuorum(n int) int {
	return n - MaxNodesFaulty(n)
}

// MinNodesFair returns a minimal number of nodes, that would include
// at least 1 fair node. It's F+1.
func MinNodesFair(n int) int {
	return MaxNodesFaulty(n) + 1
}
