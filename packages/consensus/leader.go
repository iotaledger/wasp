package consensus

import (
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

func (op *Operator) iAmCurrentLeader() bool {
	return op.committee.OwnPeerIndex() == op.currentLeaderPeerIndex()
}

func (op *Operator) currentLeaderPeerIndex() uint16 {
	return op.leaderAtSeqIndex(op.currLeaderSeqIndex)
}

func (op *Operator) leaderAtSeqIndex(seqIdx uint16) uint16 {
	return op.leaderPeerIndexList[seqIdx]
}

const leaderRotationPeriod = 3 * time.Second

func (op *Operator) moveToNextLeader() {
	op.currLeaderSeqIndex = (op.currLeaderSeqIndex + 1) % op.committee.Size()
	op.setLeaderRotationDeadline(time.Now().Add(leaderRotationPeriod))
}

func (op *Operator) resetLeader(seedBytes []byte) {
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

func (op *Operator) setLeaderRotationDeadline(deadline time.Time) {
	op.leaderRotationDeadlineSet = false
	op.leaderRotationDeadline = deadline
}

func (op *Operator) rotateLeaderIfNeeded() {
	if op.leaderRotationDeadlineSet && op.leaderRotationDeadline.After(time.Now()) {
		op.moveToNextLeader()
	}
}
