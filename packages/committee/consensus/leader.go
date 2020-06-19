package consensus

import (
	"time"
)

func (op *operator) currentLeader() (uint16, bool) {
	_, ok := op.stateIndex()
	return op.peerPermutation.Current(), ok
}

func (op *operator) iAmCurrentLeader() bool {
	idx, ok := op.currentLeader()
	return ok && op.committee.OwnPeerIndex() == idx
}

const leaderRotationPeriod = 3 * time.Second

func (op *operator) moveToNextLeader() uint16 {
	op.peerPermutation.Next()
	ret := op.moveToFirstAliveLeader()
	op.setLeaderRotationDeadline()
	return ret
}

func (op *operator) resetLeader(seedBytes []byte) {
	op.peerPermutation.Shuffle(seedBytes)
	op.leaderStatus = nil
	op.moveToFirstAliveLeader()
	op.leaderRotationDeadlineSet = false
}

// select leader first in the permutation which is alive
// then sets deadline if itself is not the leader
func (op *operator) moveToFirstAliveLeader() uint16 {
	var ret uint16
	// the loop will always stop because the current node is always alive
	for {
		if op.committee.IsAlivePeer(op.peerPermutation.Current()) {
			ret = op.peerPermutation.Current()
			break
		}
		op.log.Infof("peer #%d is dead", op.peerPermutation.Current())
		op.peerPermutation.Next()
	}
	return ret
}

func (op *operator) setLeaderRotationDeadline() {
	if len(op.requestMsgList()) == 0 {
		op.leaderRotationDeadlineSet = false
		return
	}
	op.leaderRotationDeadlineSet = true
	op.leaderRotationDeadline = time.Now().Add(leaderRotationPeriod)
}
