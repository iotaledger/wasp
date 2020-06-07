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
	ret := op.peerPermutation.Next()
	op.setLeaderRotationDeadline(time.Now().Add(leaderRotationPeriod))
	return ret
}

func (op *operator) resetLeader(seedBytes []byte) {
	op.peerPermutation.Shuffle(seedBytes)
	op.leaderStatus = nil
	op.leaderRotationDeadlineSet = false
}

func (op *operator) setLeaderRotationDeadline(deadline time.Time) {
	op.leaderRotationDeadlineSet = true
	op.leaderRotationDeadline = deadline
}

func (op *operator) rotateLeaderIfNeeded() {
	if op.leaderRotationDeadlineSet && op.leaderRotationDeadline.After(time.Now()) {
		op.moveToNextLeader()
	}
}
