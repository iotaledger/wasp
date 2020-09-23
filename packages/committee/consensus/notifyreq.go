package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/util"
)

// sendRequestNotificationsToLeader sends current leader the backlog of requests
// it is only possible in the `consensusStageLeaderStarting` stage for non-leader
func (op *operator) sendRequestNotificationsToLeader() {
	if len(op.requests) == 0 {
		return
	}
	if op.iAmCurrentLeader() {
		return
	}
	if op.consensusStage != consensusStageSubStarting {
		return
	}
	if !op.committee.HasQuorum() {
		op.log.Debugf("sendRequestNotificationsToLeader: postponed due to no quorum. Peer status: %s",
			op.committee.PeerStatus())
		return
	}
	currentLeaderPeerIndex, _ := op.currentLeader()
	reqs := op.requestCandidateList()
	reqs = op.filterOutRequestsWithoutTokens(reqs)

	// get not time-locked requests with the message known
	if len(reqs) == 0 {
		// nothing to notify about
		return
	}
	op.log.Debugf("sending notifications to #%d, backlog: %d, candidates (with tokens): %d",
		currentLeaderPeerIndex, len(op.requests), len(reqs))

	reqIds := takeIds(reqs)
	msgData := util.MustBytes(&committee.NotifyReqMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: op.mustStateIndex(),
		},
		RequestIds: reqIds,
	})

	// send until first success, but no more than number of nodes in the committee
	op.log.Infow("sendRequestNotificationsToLeader",
		"leader", currentLeaderPeerIndex,
		"state index", op.mustStateIndex(),
		"reqs", idsShortStr(reqIds),
	)
	if err := op.committee.SendMsg(currentLeaderPeerIndex, committee.MsgNotifyRequests, msgData); err != nil {
		op.log.Errorf("sending notifications to %d: %v", currentLeaderPeerIndex, err)
	}
	op.setNextConsensusStage(consensusStageSubNotificationsSent)
}

func (op *operator) storeNotification(msg *committee.NotifyReqMsg) {
	stateIndex, stateDefined := op.stateIndex()
	if stateDefined && msg.StateIndex < stateIndex {
		// don't save from earlier. The current currentSCState saved only for tracking
		return
	}
	op.notificationsBacklog = append(op.notificationsBacklog, msg)
}

// markRequestsNotified stores information about notification in the current currentSCState
func (op *operator) markRequestsNotified(msgs []*committee.NotifyReqMsg) {
	stateIndex, stateDefined := op.stateIndex()
	if !stateDefined {
		return
	}
	for _, msg := range msgs {
		if msg.StateIndex != stateIndex {
			continue
		}
		for _, reqid := range msg.RequestIds {
			req, ok := op.requestFromId(reqid)
			if !ok {
				continue
			}
			// mark request was seen by sender
			req.notifications[msg.SenderIndex] = true
		}
	}
}

// adjust all notification information to the current state index
func (op *operator) adjustNotifications() {
	stateIndex, stateDefined := op.stateIndex()
	if !stateDefined {
		return
	}
	// clear all the notification markers
	for _, req := range op.requests {
		setAllFalse(req.notifications)
		req.notifications[op.peerIndex()] = req.reqTx != nil
	}
	// put markers of the current state
	op.markRequestsNotified(op.notificationsBacklog)

	// clean notification backlog from messages from current and and past stages
	newBacklog := op.notificationsBacklog[:0] // new slice, same underlying array!
	for _, msg := range op.notificationsBacklog {
		if msg.StateIndex < stateIndex {
			continue
		}
		newBacklog = append(newBacklog, msg)
	}
	op.notificationsBacklog = newBacklog
}

func setAllFalse(bs []bool) {
	for i := range bs {
		bs[i] = false
	}
}
