package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
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
	ids := make([]*sctransaction.RequestId, 0, len(reqs))
	if len(reqs) > 0 {
		for _, r := range reqs {
			ids = append(ids, &r.reqId)
		}
	} else {
		// all of them if any
		for _, req := range op.requests {
			if req.reqTx == nil {
				continue
			}
			ids = append(ids, &req.reqId)
		}
	}
	if len(ids) == 0 {
		// nothing to notify about
		return
	}
	msgData := util.MustBytes(&committee.NotifyReqMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: op.stateTx.MustState().StateIndex(),
		},
		RequestIds: ids,
	})

	// send until first success, but no more than number of nodes in the committee
	op.log.Debugw("sendRequestNotificationsToLeader",
		"leader", currentLeaderPeerIndex,
		"my peer index", op.peerIndex(),
		"num req", len(ids),
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
		req, ok := op.requestFromId(*reqid)
		if !ok {
			continue
		}
		// mark request was seen by sender
		req.notifications[msg.SenderIndex] = true
	}
}
