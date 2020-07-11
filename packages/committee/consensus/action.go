package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"time"
)

func (op *operator) takeAction() {
	op.requestOutputsIfNeeded()
	if op.iAmCurrentLeader() {
		op.startProcessingIfNeeded()
	}
	op.checkQuorum()
	op.rotateLeaderIfNeeded()
	op.sendNotificationsOnTimeUnlock()
}

func (op *operator) sendNotificationsOnTimeUnlock() {
	numlocked := 0
	nowis := time.Now()
	reqs := make([]*request, 0)
	for _, req := range op.requests {
		if req.reqTx == nil {
			continue
		}
		if req.isTimelocked(nowis) {
			numlocked++
			continue
		}
		if !req.expectTimeUnlockEvent {
			continue
		}
		// request was just unlocked -> notifications to be sent to the leader
		reqs = append(reqs, req)
		req.expectTimeUnlockEvent = false
	}

	if len(reqs) == 0 {
		return
	}
	for _, req := range reqs {
		req.log.Infof("unlocked time lock at %d", util.TimeNowUnix())
	}
	op.sendRequestNotificationsToLeader(reqs)
}

func (op *operator) rotateLeaderIfNeeded() {
	if !op.synchronized {
		return
	}
	//if op.iAmCurrentLeader() {
	//	return
	//}
	if !op.leaderRotationDeadlineSet {
		return
	}
	if op.leaderRotationDeadline.After(time.Now()) {
		return
	}
	prevlead, _ := op.currentLeader()
	leader := op.moveToNextLeader()
	op.log.Infof("LEADER ROTATED #%d --> #%d", prevlead, leader)
	op.sendRequestNotificationsToLeader(nil)
}

func (op *operator) startProcessingIfNeeded() {
	if !op.synchronized {
		return
	}
	if op.leaderStatus != nil {
		// request already selected and calculations initialized
		return
	}

	reqs := op.selectRequestsToProcess()
	if len(reqs) == 0 {
		// can't select request to process
		//op.log.Debugf("can't select request to process")
		return
	}
	reqIds := takeIds(reqs)
	reqIdsStr := idsShortStr(reqIds)
	op.log.Debugw("requests selected to process",
		"stateIdx", op.stateTx.MustState().StateIndex(),
		"batch", reqIdsStr,
	)
	rewardAddress := op.getRewardAddress()

	// send to subordinate the request to process the batch
	msgData := util.MustBytes(&committee.StartProcessingBatchMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			// timestamp is set by SendMsgToCommitteePeers
			StateIndex: op.stateTx.MustState().StateIndex(),
		},
		RewardAddress: rewardAddress,
		Balances:      op.balances,
		RequestIds:    reqIds,
	})

	numSucc, ts := op.committee.SendMsgToCommitteePeers(committee.MsgStartProcessingRequest, msgData)

	op.log.Debugf("%d 'msgStartProcessingRequest' messages sent to peers", numSucc)

	if numSucc < op.quorum()-1 {
		// doesn't make sense to continue because less than quorum sends succeeded
		op.log.Errorf("only %d 'msgStartProcessingRequest' sends succeeded.", numSucc)
		return
	}

	batchHash := vm.BatchHash(reqIds, ts, op.peerIndex())
	op.leaderStatus = &leaderStatus{
		reqs:          reqs,
		batchHash:     batchHash,
		balances:      op.balances,
		timestamp:     ts,
		signedResults: make([]*signedResult, op.committee.Size()),
	}
	op.log.Debugw("runCalculationsAsync leader",
		"batch hash", batchHash.String(),
		"batch", reqIdsStr,
		"ts", ts,
	)
	// process the batch on own side
	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		leaderPeerIndex: op.committee.OwnPeerIndex(),
		balances:        op.balances,
		timestamp:       ts,
		rewardAddress:   rewardAddress,
	})
}

func (op *operator) checkQuorum() bool {
	if !op.synchronized {
		return false
	}
	if op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized {
		return false
	}

	// collect signature shares available
	mainHash := op.leaderStatus.signedResults[op.committee.OwnPeerIndex()].essenceHash
	sigShares := make([][]byte, 0, op.committee.Size())
	contributingPeers := make([]uint16, 0, op.size())
	for i := range op.leaderStatus.signedResults {
		if op.leaderStatus.signedResults[i] == nil {
			continue
		}
		if op.leaderStatus.signedResults[i].essenceHash != mainHash {
			op.log.Warnf("wrong EssenceHash from peer #%d: %s",
				i, op.leaderStatus.signedResults[i].essenceHash.String())
			op.leaderStatus.signedResults[i] = nil // ignoring
			continue
		}
		err := op.dkshare.VerifySigShare(op.leaderStatus.resultTx.EssenceBytes(), op.leaderStatus.signedResults[i].sigShare)
		if err != nil {
			op.log.Warnf("wrong signature from peer #%d: %v", i, err)
			op.leaderStatus.signedResults[i] = nil // ignoring
			continue
		}

		sigShare := op.leaderStatus.signedResults[i].sigShare
		idx, _ := sigShare.Index()
		sigShares = append(sigShares, sigShare)
		contributingPeers = append(contributingPeers, uint16(idx))
	}

	if len(sigShares) < int(op.quorum()) {
		return false
	}
	// quorum detected
	finalSignature, err := op.aggregateSigShares(sigShares)
	if err != nil {
		op.log.Errorf("aggregateSigShares returned: %v", err)
		return false
	}

	if !op.leaderStatus.resultTx.SignaturesValid() {
		op.log.Error("final signature invalid: something went wrong while finalizing result transaction")
		return false
	}

	sh := op.leaderStatus.resultTx.MustState().StateHash()
	stateIndex := op.leaderStatus.resultTx.MustState().StateIndex()
	op.log.Infof("FINALIZED RESULT. txid: %s, state index: #%d, state hash: %s, contributors: %+v",
		op.leaderStatus.resultTx.ID().String(), stateIndex, sh.String(), contributingPeers)
	op.leaderStatus.finalized = true

	if err = nodeconn.PostTransactionToNode(op.leaderStatus.resultTx.Transaction); err != nil {
		op.log.Warnf("PostTransactionToNode failed: %v", err)
		return false
	}

	// notify peers about finalization
	msgData := util.MustBytes(&committee.NotifyFinalResultPostedMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			// timestamp is set by SendMsgToCommitteePeers
			StateIndex: op.stateTx.MustState().StateIndex(),
		},
		Signature: finalSignature,
	})

	op.committee.SendMsgToCommitteePeers(committee.MsgNotifyFinalResultPosted, msgData)
	return true
}

// sets new currentState transaction and initializes respective variables
func (op *operator) setNewState(stateTx *sctransaction.Transaction, variableState state.VirtualState, synchronized bool) {
	op.stateTx = stateTx
	op.currentState = variableState
	op.synchronized = synchronized

	op.requestBalancesDeadline = time.Now()
	op.requestOutputsIfNeeded()

	op.resetLeader(stateTx.ID().Bytes())

	op.adjustNotifications()
}

func (op *operator) requestOutputsIfNeeded() {
	if !op.synchronized {
		return
	}
	if op.balances != nil && op.requestBalancesDeadline.After(time.Now()) {
		return
	}
	if err := nodeconn.RequestOutputsFromNode(op.committee.Address()); err != nil {
		op.log.Debugf("RequestOutputsFromNode failed: %v", err)
	}
	op.requestBalancesDeadline = time.Now().Add(committee.RequestBalancesPeriod)
}
