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

func (op *operator) moveToNextLeader() uint16 {
	op.peerPermutation.Next()
	ret := op.moveToFirstAliveLeader()
	//op.setLeaderRotationDeadline(op.committee.Params().LeaderReactionToNotifications)
	return ret
}

func (op *operator) resetLeader(seedBytes []byte) {
	op.peerPermutation.Shuffle(seedBytes)
	op.leaderStatus = nil
	leader := op.peerPermutation.Current()
	leader = op.moveToFirstAliveLeader()
	op.leaderRotationDeadlineSet = false
	op.stateTxEvidenced = false

	op.log.Debugf("peerPermutation: %+v, first alive leader: %d", op.peerPermutation.GetArray(), leader)
}

// select leader first in the permutation which is alive
// then sets deadline if itself is not the leader
func (op *operator) moveToFirstAliveLeader() uint16 {
	if !op.committee.HasQuorum() {
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
	// should not come here
	return op.peerPermutation.Current()
}

func (op *operator) setLeaderRotationDeadline(period time.Duration) {
	if len(op.requestCandidateList()) == 0 {
		op.leaderRotationDeadlineSet = false
		op.stateTxEvidenced = false

		op.log.Info("delete leader rotation deadline")
		return
	}
	if op.leaderRotationDeadlineSet && op.leaderRotationDeadline.After(time.Now().Add(period)) {
		// only move deadline further, not back
		return
	}
	op.leaderRotationDeadlineSet = true
	op.leaderRotationDeadline = time.Now().Add(period)

	op.log.Infof("set leader rotation deadline to %v", period)
}
