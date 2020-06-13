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
	ret := op.selectFistAliveLeader()
	op.setLeaderRotationDeadline(time.Now().Add(leaderRotationPeriod))
	return ret
}

func (op *operator) resetLeader(seedBytes []byte) {
	op.peerPermutation.Shuffle(seedBytes)
	op.leaderStatus = nil
	op.leaderRotationDeadlineSet = false
	op.selectFistAliveLeader()
}

// select leader first in the permutation which is alive
func (op *operator) selectFistAliveLeader() uint16 {
	// the loop will always stop because the current node is always alive
	for {
		if op.committee.IsAlivePeer(op.peerPermutation.Current()) {
			return op.peerPermutation.Current()
		}
		op.peerPermutation.Next()
	}
}

func (op *operator) setLeaderRotationDeadline(deadline time.Time) {
	op.leaderRotationDeadlineSet = !op.iAmCurrentLeader()
	if op.leaderRotationDeadlineSet {
		op.leaderRotationDeadline = deadline
	}
}
