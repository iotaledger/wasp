// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensusimpl_old

func (op *operator) currentLeader() (uint16, bool) {
	_, ok := op.blockIndex()
	return op.peerPermutation.Current(), ok
}

func (op *operator) iAmCurrentLeader() bool {
	idx, ok := op.currentLeader()
	return ok && op.committee.OwnPeerIndex() == idx
}

func (op *operator) moveToNextLeader() uint16 {
	op.peerPermutation.Next()
	ret := op.moveToFirstAliveLeader()
	return ret
}

func (op *operator) resetLeader(seedBytes []byte) {
	op.peerPermutation.Shuffle(seedBytes)
	op.leaderStatus = nil
	leader := op.moveToFirstAliveLeader()

	op.log.Debugf("peerPermutation: %+v, leader: %d", op.peerPermutation.GetArray(), leader)
}

// select leader first in the permutation which is alive
// then sets deadline if itself is not the leader
func (op *operator) moveToFirstAliveLeader() uint16 {
	if !op.committee.QuorumIsAlive() {
		// not enough alive nodes, just do nothing
		return op.peerPermutation.Current()
	}
	// the loop will always stop because the current node is always alive
	for {
		if op.committee.IsAlivePeer(op.peerPermutation.Current()) {
			break
		}
		op.log.Debugf("peer #%d is not alive", op.peerPermutation.Current())
		op.peerPermutation.Next()
	}
	return op.peerPermutation.Current()
}
