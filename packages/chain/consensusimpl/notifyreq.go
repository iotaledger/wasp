// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensusimpl

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/util"
)

// sendRequestNotificationsToLeader sends current leader the backlog of requests
// it is only possible in the `consensusStageLeaderStarting` stage for non-leader
func (op *operator) sendRequestNotificationsToLeader() {
	if op.consensusStage != consensusStageSubStarting {
		return
	}
	stateIndex, stateExist := op.blockIndex()
	if !stateExist {
		return
	}
	if op.iAmCurrentLeader() {
		return
	}
	readyRequests := op.mempool.GetReadyList()
	if len(readyRequests) == 0 {
		return
	}
	reqIds := takeIDs(readyRequests...)
	currentLeaderPeerIndex, _ := op.currentLeader()

	op.log.Debugf("sending request notifications to #%d: %+v", currentLeaderPeerIndex, idsShortStr(reqIds...))

	msgData := util.MustBytes(&chain.NotifyReqMsg{
		StateOutputID: op.stateOutput.ID(),
		RequestIDs:    reqIds,
	})

	// send until first success, but no more than number of nodes in the committee
	op.log.Infow("sendRequestNotificationsToLeader",
		"leader", currentLeaderPeerIndex,
		"state index", stateIndex,
		"reqs", idsShortStr(reqIds...),
	)
	if err := op.committee.SendMsg(currentLeaderPeerIndex, chain.MsgNotifyRequests, msgData); err != nil {
		op.log.Errorf("sending notifications to %d: %v", currentLeaderPeerIndex, err)
	}
	op.setNextConsensusStage(consensusStageSubNotificationsSent)
}
