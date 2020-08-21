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
	op.sendRequestNotificationsToLeader()
	op.queryOutputs()
	op.startCalculations()
	op.checkQuorum()
	op.rotateLeader()
}

// rotateLeader upon expired deadline
func (op *operator) rotateLeader() {
	if !op.consensusStageDeadlineExpired() {
		return
	}
	if !op.committee.HasQuorum() {
		op.log.Debugf("leader not rotated due to no quorum")
		return
	}
	prevlead, _ := op.currentLeader()
	leader := op.moveToNextLeader()

	// starting from scratch with the new leader
	op.leaderStatus = nil
	op.sentResultToLeader = nil
	op.setConsensusStage(consensusStageLeaderStarting)

	op.log.Infof("LEADER ROTATED #%d --> #%d, I am the leader = %v",
		prevlead, leader, op.iAmCurrentLeader())
}

func (op *operator) startCalculations() {
	if op.consensusStage != consensusStageLeaderStarting || !op.iAmCurrentLeader() {
		// only for leader in the beginning of the starting stage
		return
	}
	if !op.committee.HasQuorum() {
		// no quorum, doesn't make sense to start
		return
	}
	// select requests for the batch
	reqs := op.selectRequestsToProcess()
	if len(reqs) == 0 {
		return
	}
	reqIds := takeIds(reqs)
	reqIdsStr := idsShortStr(reqIds)
	op.log.Debugf("requests selected to process. State: %d, Reqs: %+v",
		op.stateTx.MustState().StateIndex(), reqIdsStr,
	)
	rewardAddress := op.getRewardAddress()

	// send to subordinated peers requests to process the batch
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
		op.log.Errorf("only %d 'msgStartProcessingRequest' sends succeeded. Not continuing", numSucc)
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
	op.setConsensusStage(consensusStageCalculationsStarted)
}

func (op *operator) checkQuorum() bool {
	if !op.iAmCurrentLeader() {
		return false
	}
	if op.consensusStage != consensusStageCalculationsFinished {
		// checking quorum only if own calculations has been finished
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
	op.log.Infof("FINALIZED RESULT. txid: %s, consensusStage index: #%d, consensusStage hash: %s, contributors: %+v",
		op.leaderStatus.resultTx.ID().String(), stateIndex, sh.String(), contributingPeers)
	op.leaderStatus.finalized = true

	err = nodeconn.PostTransactionToNode(op.leaderStatus.resultTx.Transaction, op.committee.Address(), op.committee.OwnPeerIndex())
	if err != nil {
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
	op.setConsensusStage(consensusStageResultFinalized)
	return true
}

// sets new currentSCState transaction and initializes respective variables
func (op *operator) setNewSCState(stateTx *sctransaction.Transaction, variableState state.VirtualState, synchronized bool) {
	op.stateTx = stateTx
	op.currentSCState = variableState
	op.sentResultToLeader = nil

	op.requestBalancesDeadline = time.Now()
	op.queryOutputs()

	op.resetLeader(stateTx.ID().Bytes())

	op.adjustNotifications()
}

func (op *operator) queryOutputs() {
	if op.consensusStage != consensusStageNoSync {
		return
	}
	if op.balances != nil && op.requestBalancesDeadline.After(time.Now()) {
		return
	}
	if err := nodeconn.RequestOutputsFromNode(op.committee.Address()); err != nil {
		op.log.Debugf("RequestOutputsFromNode failed: %v", err)
	}
	op.requestBalancesDeadline = time.Now().Add(op.committee.Params().RequestBalancesPeriod)
}
