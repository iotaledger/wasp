package consensus

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/plugins/nodeconn"
)

func (op *operator) takeAction() {
	op.sendRequestNotificationsToLeader()
	//op.queryOutputs()
	op.startCalculationsAsLeader()
	op.checkQuorum()
	op.rotateLeader()
	op.pullInclusionLevel()
}

func (op *operator) pullInclusionLevel() {
	if op.postedResultTxid == nil {
		return
	}
	if time.Now().After(op.nextPullInclusionLevel) {
		addr := address.Address(*op.chain.ID())
		if err := nodeconn.RequestInclusionLevelFromNode(op.postedResultTxid, &addr); err != nil {
			op.log.Errorf("RequestInclusionLevelFromNode: %v", err)
		}
		op.setNextPullInclusionStageDeadline()
	}
}

// rotateLeader upon expired deadline
func (op *operator) rotateLeader() {
	if !op.consensusStageDeadlineExpired() {
		return
	}
	if !op.chain.HasQuorum() {
		op.log.Debugf("leader not rotated due to no quorum")
		return
	}
	prevlead, _ := op.currentLeader()
	leader := op.moveToNextLeader()

	// starting from scratch with the new leader
	op.leaderStatus = nil
	op.sentResultToLeader = nil
	op.postedResultTxid = nil

	op.log.Infof("LEADER ROTATED #%d --> #%d, I am the leader = %v",
		prevlead, leader, op.iAmCurrentLeader())

	if op.iAmCurrentLeader() {
		op.setNextConsensusStage(consensusStageLeaderStarting)
	} else {
		op.setNextConsensusStage(consensusStageSubStarting)
	}
}

func (op *operator) startCalculationsAsLeader() {
	if op.consensusStage != consensusStageLeaderStarting {
		// only for leader in the beginning of the starting stage
		return
	}
	if !op.chain.HasQuorum() {
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

	op.log.Debugf("requests selected to process. Current state: %d, Reqs: %+v", op.mustStateIndex(), reqIdsStr)
	rewardAddress := op.getRewardAddress()

	// send to subordinated peers requests to process the batch
	msgData := util.MustBytes(&chain.StartProcessingBatchMsg{
		PeerMsgHeader: chain.PeerMsgHeader{
			// timestamp is set by SendMsgToCommitteePeers
			BlockIndex: op.stateTx.MustState().StateIndex(),
		},
		RewardAddress: rewardAddress,
		Balances:      op.balances,
		RequestIds:    reqIds,
	})

	// determine timestamp. Must be max(local clock, prev timestamp+1)
	ts := time.Now().UnixNano()
	prevTs := op.stateTx.MustState().Timestamp()
	if ts <= prevTs {
		op.log.Warnf("local clock is not ahead the timestamp of the previous state. prevTs: %d, currentTs: %d, diff: %d ns",
			prevTs, ts, prevTs-ts)
		ts = prevTs + 1
		op.log.Info("timestamp was adjusted to %d", ts)
	}

	numSucc := op.chain.SendMsgToCommitteePeers(chain.MsgStartProcessingRequest, msgData, ts)

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
		signedResults: make([]*signedResult, op.chain.Size()),
	}
	op.log.Debugw("runCalculationsAsync leader",
		"batch hash", batchHash.String(),
		"batch", reqIdsStr,
		"ts", ts,
	)
	// process the batch on own side
	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		leaderPeerIndex: op.chain.OwnPeerIndex(),
		balances:        op.balances,
		timestamp:       ts,
		rewardAddress:   rewardAddress,
	})
	op.setNextConsensusStage(consensusStageLeaderCalculationsStarted)
}

func (op *operator) checkQuorum() bool {
	if op.consensusStage != consensusStageLeaderCalculationsFinished {
		// checking quorum only if leader calculations has been finished
		return false
	}
	if op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized {
		return false
	}

	// collect signature shares available
	mainHash := op.leaderStatus.signedResults[op.chain.OwnPeerIndex()].essenceHash
	sigShares := make([][]byte, 0, op.chain.Size())
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
	if err := op.aggregateSigShares(sigShares); err != nil {
		op.log.Errorf("aggregateSigShares returned: %v", err)
		return false
	}

	if !op.leaderStatus.resultTx.SignaturesValid() {
		op.log.Error("final signature invalid: something went wrong while finalizing result transaction")
		return false
	}

	txid := op.leaderStatus.resultTx.ID()
	sh := op.leaderStatus.resultTx.MustState().StateHash()
	stateIndex := op.leaderStatus.resultTx.MustState().StateIndex()
	op.log.Infof("FINALIZED RESULT. txid: %s, state index: #%d, state hash: %s, contributors: %+v",
		txid.String(), stateIndex, sh.String(), contributingPeers)
	op.leaderStatus.finalized = true

	addr := address.Address(*op.chain.ID())
	err := nodeconn.PostTransactionToNode(op.leaderStatus.resultTx.Transaction, &addr, op.chain.OwnPeerIndex())
	if err != nil {
		op.log.Warnf("PostTransactionToNode failed: %v", err)
		return false
	}
	op.log.Debugf("result transaction has been posted to node. txid: %s", txid.String())

	// notify peers about finalization
	msgData := util.MustBytes(&chain.NotifyFinalResultPostedMsg{
		PeerMsgHeader: chain.PeerMsgHeader{
			// timestamp is set by SendMsgToCommitteePeers
			BlockIndex: op.stateTx.MustState().StateIndex(),
		},
		TxId: txid,
	})

	numSent := op.chain.SendMsgToCommitteePeers(chain.MsgNotifyFinalResultPosted, msgData, time.Now().UnixNano())
	op.log.Debugf("%d peers has been notified about finalized result", numSent)

	op.setNextConsensusStage(consensusStageLeaderResultFinalized)
	op.setFinalizedTransaction(&txid)

	return true
}

// sets new currentState transaction and initializes respective variables
func (op *operator) setNewSCState(stateTx *sctransaction.Transaction, variableState state.VirtualState, synchronized bool) {
	op.stateTx = stateTx
	op.currentState = variableState
	op.sentResultToLeader = nil
	op.postedResultTxid = nil

	op.requestBalancesDeadline = time.Now()
	//op.queryOutputs()

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
	addr := address.Address(*op.chain.ID())
	if err := nodeconn.RequestOutputsFromNode(&addr); err != nil {
		op.log.Debugf("RequestOutputsFromNode failed: %v", err)
	}
	op.requestBalancesDeadline = time.Now().Add(chain.RequestBalancesPeriod)
}
