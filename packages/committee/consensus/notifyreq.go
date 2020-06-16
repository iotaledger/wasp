package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/util"
)

// notifies current leader about requests in the order of arrival
func (op *operator) sendRequestNotificationsToLeader(reqs []*request) {
	stateIndex, ok := op.stateIndex()
	if !ok {
		op.log.Debugf("sendRequestNotificationsToLeader: current state is undefined")
		return
	}
	currentLeaderPeerIndex, ok := op.currentLeader()
	if !ok {
		return
	}
	if op.iAmCurrentLeader() {
		return
	}
	if reqs == nil {
		reqs = op.requestMsgList()
	}
	reqIds := takeIds(reqs)
	if len(reqIds) == 0 {
		// nothing to notify about
		return
	}
	msgData := util.MustBytes(&committee.NotifyReqMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: stateIndex,
		},
		RequestIds: reqIds,
	})

	// send until first success, but no more than number of nodes in the committee
	op.log.Debugw("sendRequestNotificationsToLeader",
		"leader", currentLeaderPeerIndex,
		"state index", stateIndex,
		"reqs", idsShortStr(reqIds),
	)

	if err := op.committee.SendMsg(currentLeaderPeerIndex, committee.MsgNotifyRequests, msgData); err != nil {
		op.log.Debugf("sending notifications to %d: %v", currentLeaderPeerIndex, err)
	}
}

func (op *operator) storeNotificationIfNeeded(msg *committee.NotifyReqMsg) {
	stateIndex, stateDefined := op.stateIndex()
	if stateDefined && msg.StateIndex <= stateIndex {
		// don't save from the current state and earlier
		return
	}
	op.notificationsBacklog = append(op.notificationsBacklog, msg)
}

// stores information about notification in the current state
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

	// clean notification backlog from messages from current and and past states
	newBacklog := op.notificationsBacklog[:0] // new slice, same underlying array!
	for _, msg := range op.notificationsBacklog {
		if msg.StateIndex <= stateIndex {
			continue
		}
		newBacklog = append(newBacklog, msg)
	}
	op.notificationsBacklog = newBacklog
}
