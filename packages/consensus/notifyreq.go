package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/util"
)

// notifies current leader about requests in the order of arrival
func (op *operator) sendRequestNotificationsToLeader(reqs []*request) {
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
			StateIndex: op.stateIndex(),
		},
		RequestIds: reqIds,
	})

	// send until first success, but no more than number of nodes in the committee
	op.log.Debugw("sendRequestNotificationsToLeader",
		"state index", op.stateIndex(),
		"leader", currentLeaderPeerIndex,
		"reqs", idsShortStr(reqIds),
	)

	if err := op.committee.SendMsg(currentLeaderPeerIndex, committee.MsgNotifyRequests, msgData); err != nil {
		op.log.Debugf("sending notifications to %d: %v", currentLeaderPeerIndex, err)
	}
}

// stores information about notification in the current state
func (op *operator) markRequestsNotified(msg *committee.NotifyReqMsg) {
	if op.variableState == nil {
		op.notificationsBacklog = append(op.notificationsBacklog, msg)
		return
	}
	if msg.StateIndex < op.variableState.StateIndex() {
		// ignore
		return
	}
	if msg.StateIndex > op.variableState.StateIndex() {
		op.notificationsBacklog = append(op.notificationsBacklog, msg)
		return
	}
	if !(op.variableState != nil && msg.StateIndex == op.variableState.StateIndex()) {
		panic("assertion failed: op.variableState != nil && msg.StateIndex == op.variableState.StateIndex()")
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
