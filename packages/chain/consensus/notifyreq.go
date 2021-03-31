// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
)

// sendRequestNotificationsToLeader sends current leader the backlog of requests
// it is only possible in the `consensusStageLeaderStarting` stage for non-leader
func (op *operator) sendRequestNotificationsToLeader() {
	readyRequests := op.mempool.GetReadyList(op.quorum())
	reqIds := make([]coretypes.RequestID, len(readyRequests))
	for i := range reqIds {
		reqIds[i] = readyRequests[i].ID()
	}
	currentLeaderPeerIndex, _ := op.currentLeader()

	op.log.Debugf("sending %d request notifications to #%d", len(readyRequests), currentLeaderPeerIndex)

	msgData := util.MustBytes(&chain.NotifyReqMsg{
		PeerMsgHeader: chain.PeerMsgHeader{
			BlockIndex: op.mustStateIndex(),
		},
		RequestIDs: reqIds,
	})

	// send until first success, but no more than number of nodes in the committee
	op.log.Infow("sendRequestNotificationsToLeader",
		"leader", currentLeaderPeerIndex,
		"state index", op.mustStateIndex(),
		"reqs", idsShortStr(reqIds),
	)
	if err := op.committee.SendMsg(currentLeaderPeerIndex, chain.MsgNotifyRequests, msgData); err != nil {
		op.log.Errorf("sending notifications to %d: %v", currentLeaderPeerIndex, err)
	}
	op.setNextConsensusStage(consensusStageSubNotificationsSent)
}

// adjust all notification information to the current state index
func (op *operator) adjustNotifications() {
	stateIndex, stateDefined := op.blockIndex()
	if !stateDefined {
		return
	}
	// clear all the notification markers
	for _, req := range op.requests {
		setAllFalse(req.notifications)
		req.notifications[op.peerIndex()] = req.req != nil
	}
	// put markers of the current state
	op.markRequestsNotified(op.notificationsBacklog)

	// clean notification backlog from messages from current and and past stages
	newBacklog := op.notificationsBacklog[:0] // new slice, same underlying array!
	for _, msg := range op.notificationsBacklog {
		if msg.BlockIndex < stateIndex {
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
