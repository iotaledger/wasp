package consensus

import (
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

func (op *operator) iAmCurrentLeader() bool {
	currLead, ok := op.currentLeaderPeerIndex()
	if !ok {
		return false
	}
	return op.committee.OwnPeerIndex() == currLead
}

func (op *operator) currentLeaderPeerIndex() (uint16, bool) {
	return op.leaderAtSeqIndex(op.currLeaderSeqIndex)
}

func (op *operator) leaderAtSeqIndex(seqIdx uint16) (uint16, bool) {
	if op.leaderPeerIndexList == nil {
		return 0, false
	}
	return op.leaderPeerIndexList[seqIdx], true
}

const leaderRotationPeriod = 3 * time.Second

func (op *operator) moveToNextLeader() {
	op.currLeaderSeqIndex = (op.currLeaderSeqIndex + 1) % op.committee.Size()
	op.setLeaderRotationDeadline(time.Now().Add(leaderRotationPeriod))
}

func (op *operator) resetLeader(seedBytes []byte) {
	op.currLeaderSeqIndex = 0
	op.leaderPeerIndexList = util.GetPermutation(op.committee.Size(), seedBytes)
	for i, v := range op.leaderPeerIndexList {
		if v == op.committee.OwnPeerIndex() {
			op.myLeaderSeqIndex = uint16(i)
			break
		}
	}
	// leader part of processing wasn't started yet
	op.leaderStatus = nil
	op.leaderRotationDeadlineSet = false
}

func (op *operator) setLeaderRotationDeadline(deadline time.Time) {
	op.leaderRotationDeadlineSet = false
	op.leaderRotationDeadline = deadline
}

func (op *operator) rotateLeaderIfNeeded() {
	if op.leaderRotationDeadlineSet && op.leaderRotationDeadline.After(time.Now()) {
		op.moveToNextLeader()
	}
}
