package consensus

import (
	"time"
)

func (op *operator) currentLeader() (uint16, bool) {
	if op.stateTx == nil {
		return 0, false
	}
	return op.peerPermutation.Current(), true
}

func (op *operator) iAmCurrentLeader() bool {
	idx, ok := op.currentLeader()
	return ok && op.committee.OwnPeerIndex() == idx
}

const leaderRotationPeriod = 3 * time.Second

func (op *operator) moveToNextLeader() uint16 {
	op.peerPermutation.Next()
	return op.moveToFirstAliveLeader()
}

func (op *operator) resetLeader(seedBytes []byte) {
	op.peerPermutation.Shuffle(seedBytes)
	op.leaderStatus = nil
	op.moveToFirstAliveLeader()
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
		op.log.Debugf("peer #%d is dead", op.peerPermutation.Current())
		op.peerPermutation.Next()
	}
	op.setLeaderRotationDeadline()
	return ret
}

func (op *operator) setLeaderRotationDeadline() {
	if len(op.requestMsgList()) == 0 || op.iAmCurrentLeader() {
		op.leaderRotationDeadlineSet = false
		return
	}
	if !op.leaderRotationDeadlineSet {
		op.leaderRotationDeadlineSet = true
		op.leaderRotationDeadline = time.Now().Add(leaderRotationPeriod)
	}
}
