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

// stores information about notification in the current state
func (op *operator) markRequestsNotified(msgs []*committee.NotifyReqMsg) {
	stateIndex, stateDefined := op.stateIndex()
	for _, msg := range msgs {
		if !stateDefined && msg.StateIndex != 0 {
			continue
		}
		if stateDefined && msg.StateIndex != stateIndex {
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
