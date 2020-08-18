package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/util"
)

// sendRequestNotificationsToLeaderIfNeeded sends current leader the backlog of requests
func (op *operator) sendRequestNotificationsToLeaderIfNeeded() {
	if !op.sendNotificationsScheduled {
		return
	}
	op.sendNotificationsScheduled = false

	stateIndex, ok := op.stateIndex()
	if !ok {
		op.log.Debugf("sendRequestNotificationsToLeaderIfNeeded: current state is undefined")
		return
	}
	if !op.committee.HasQuorum() {
		op.log.Debugf("sendRequestNotificationsToLeaderIfNeeded: postponed due to no quorum")
		op.sendNotificationsScheduled = true
		return
	}

	currentLeaderPeerIndex, ok := op.currentLeader()
	if !ok {
		return
	}

	if op.iAmCurrentLeader() {
		return
	}
	op.log.Debugf("sendRequestNotificationsToLeaderIfNeeded #%d", currentLeaderPeerIndex)

	// get not time-locked requests with the message known
	reqs := op.requestCandidateList()
	if len(reqs) == 0 {
		// nothing to notify about
		return
	}
	reqIds := takeIds(reqs)
	msgData := util.MustBytes(&committee.NotifyReqMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: stateIndex,
		},
		RequestIds: reqIds,
	})

	// send until first success, but no more than number of nodes in the committee
	op.log.Infow("sendRequestNotificationsToLeaderIfNeeded",
		"leader", currentLeaderPeerIndex,
		"currentState index", stateIndex,
		"reqs", idsShortStr(reqIds),
	)
	if err := op.committee.SendMsg(currentLeaderPeerIndex, committee.MsgNotifyRequests, msgData); err != nil {
		op.log.Infof("sending notifications to %d: %v", currentLeaderPeerIndex, err)
	}
	op.setLeaderRotationDeadline(op.committee.Params().LeaderReactionToNotifications)
}

func (op *operator) storeNotificationIfNeeded(msg *committee.NotifyReqMsg) {
	stateIndex, stateDefined := op.stateIndex()
	if stateDefined && msg.StateIndex < stateIndex {
		// don't save from earlier. The current currentState saved only for tracking
		return
	}
	op.notificationsBacklog = append(op.notificationsBacklog, msg)
}

// stores information about notification in the current currentState
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

// adjust all notification information to the current currentState index
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
	// put markers of the current currentState
	op.markRequestsNotified(op.notificationsBacklog)

	// clean notification backlog from messages from current and and past states
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
