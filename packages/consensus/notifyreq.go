package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"time"
)

// notifies current leader about requests in the order of arrival
func (op *operator) sendRequestNotificationsToLeader(reqs []*request) {
	if op.iAmCurrentLeader() {
		return
	}
	ids := make([]*sctransaction.RequestId, len(reqs))
	for i := range ids {
		ids[i] = reqs[i].reqId
	}
	msgData := hashing.MustBytes(&committee.NotifyReqMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: op.stateTx.MustState().StateIndex(),
		},
		RequestIds: ids,
	})
	// send until first success, but no more than number of nodes in the committee
	var i uint16
	for i = 0; i < op.committee.Size(); i++ {
		if op.iAmCurrentLeader() {
			// stop if I am the current leader
			return
		}
		if !op.committee.IsAlivePeer(op.currentLeaderPeerIndex()) {
			op.moveToNextLeader()
			continue
		}
		err := op.committee.SendMsg(op.currentLeaderPeerIndex(), committee.MsgNotifyRequests, msgData)
		if err == nil {
			op.setLeaderRotationDeadline(time.Now().Add(leaderRotationPeriod))
			// first node to which data was successfully sent is assumed the leader
			return
		}
	}
}

// only requests with reqRef != nil
func (op *operator) sortedRequestsByAge() []*request {
	//ret := make([]*request, 0, len(op.requests))
	//for _, reqs := range op.requests {
	//	if reqs.reqRef != nil {
	//		ret = append(ret, reqs)
	//	}
	//}
	//sortRequestsByAge(ret)
	//return ret
	panic("implement me")
}

func (op *operator) sortedRequestIdsByAge() []*sctransaction.RequestId {
	sortedReqs := op.sortedRequestsByAge()
	ids := make([]*sctransaction.RequestId, len(sortedReqs))
	for i := range ids {
		ids[i] = sortedReqs[i].reqId
	}
	return ids
}

// includes request ids into the respective list of notifications,
// by the sender index
func (op *operator) markRequestsNotified(senderIndex uint16, stateIndex uint32, reqs []*sctransaction.RequestId) {
	var isCurrentState bool
	switch {
	case stateIndex == op.stateIndex():
		isCurrentState = true
	case stateIndex == op.stateIndex()+1:
		isCurrentState = false
	default:
		// from another state
		return
	}
	for _, reqid := range reqs {
		req, ok := op.requestFromId(reqid)
		if !ok {
			continue
		}
		if isCurrentState {
			req.markSeenCurrentStateBy(senderIndex)
		} else {
			req.markSeenNextStateBy(senderIndex)
		}
	}
}
