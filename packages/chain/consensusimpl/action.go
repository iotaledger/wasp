// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensusimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/parameters"
	"golang.org/x/xerrors"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
)

// takeAction analyzes the state and updates it and takes action such as sending of message,
// IF IT IS REQUIRED BY THE STATE (for example if deadline achieved, if needed data is not here and similar .
// Is called from timer ticks, also when messages received
func (op *operator) takeAction() {
	op.sendRequestNotificationsToLeader()
	op.startCalculationsAsLeader()
	op.checkQuorum()
	op.rotateLeader()
	op.pullInclusionLevel()
	op.pullBacklog()
}

func (op *operator) pullBacklog() {
	if time.Now().After(op.pullBacklogDeadline) {
		op.nodeConn.PullBacklog(op.chain.ID().AsAliasAddress())
		op.pullBacklogDeadline = time.Now().Add(pullBacklogPeriod)
	}
}

// pullInclusionLevel if it is known that result transaction was posted by the leader,
// some updates from Goshimmer are expected about the status (inclusion level) of the transaction
// If the update about the tx state didn't come as expected (timeout), send the query about it
// to the goshimmer (pull)
func (op *operator) pullInclusionLevel() {
	if op.postedResultTxid == nilTxID {
		return
	}
	if time.Now().After(op.nextPullInclusionLevel) {
		op.nodeConn.PullTransactionInclusionState(op.chain.ID().AsAddress(), op.postedResultTxid)
		op.setNextPullInclusionStageDeadline()
	}
}

// rotateLeader upon expired deadline. The deadline depends on the stage
func (op *operator) rotateLeader() {
	if !op.consensusStageDeadlineExpired() {
		return
	}
	if !op.committee.QuorumIsAlive() {
		op.log.Debugf("leader was not rotated due to no quorum")
		return
	}
	prevlead, _ := op.currentLeader()
	leader := op.moveToNextLeader()

	// starting from scratch with the new leader
	op.leaderStatus = nil
	op.sentResultToLeader = nil
	op.postedResultTxid = nilTxID

	op.log.Infof("LEADER ROTATED #%d --> #%d, I am the leader = %v",
		prevlead, leader, op.iAmCurrentLeader())

	// the consensus stage will become one of two, depending is iAmLeader or not
	if op.iAmCurrentLeader() {
		op.setNextConsensusStage(consensusStageLeaderStarting)
	} else {
		op.setNextConsensusStage(consensusStageSubStarting)
	}
}

// startCalculationsAsLeader starts calculation at the leader side at the
// 'leaderStarting' stage
func (op *operator) startCalculationsAsLeader() {
	if op.consensusStage != consensusStageLeaderStarting {
		// only for leader in the beginning of the starting stage
		return
	}
	if !op.committee.QuorumIsAlive() {
		op.log.Debugf("no quorum")
		// no quorum, doesn't make sense to start
		return
	}
	// select requests for the batch
	reqs := op.selectRequestsToProcess()
	if len(reqs) == 0 {
		// empty backlog or nothing is ready
		//op.log.Debugf("empty backlog")
		return
	}
	reqIds := takeIDs(reqs...)
	reqIdsStr := idsShortStr(reqIds...)

	op.log.Debugf("requests selected to process. Current state: %d, Reqs: %+v", op.mustStateIndex(), reqIdsStr)
	rewardAddress := op.getFeeDestination()

	// send to subordinated peers requests to process the batch
	msgData := util.MustBytes(&chain.StartProcessingBatchMsg{
		StateOutputID:  op.stateOutput.ID(),
		FeeDestination: rewardAddress,
		RequestIDs:     reqIds,
	})

	// determine timestamp. Must be max(local clock, prev timestamp+1).
	// Adjustment enforced, when needed
	ts := time.Now()
	prevTs := op.stateTimestamp
	if !ts.After(prevTs) {
		op.log.Warn("local clock is not ahead of the timestamp of the previous state")
		ts = prevTs.Add(1 * time.Nanosecond)
		op.log.Info("timestamp was adjusted to %v", ts)
	}

	numSucc := op.committee.SendMsgToPeers(chain.MsgStartProcessingRequest, msgData, ts.UnixNano())

	op.log.Debugf("%d 'msgStartProcessingRequest' messages sent to peers", numSucc)

	if numSucc < op.quorum()-1 {
		// doesn't make sense to continue with own calculations when less than quorum sends succeeds
		// should not happen normally
		op.log.Errorf("only %d 'msgStartProcessingRequest' sends succeeded. Not continuing", numSucc)
		return
	}
	// batchHash uniquely identifies inputs to calculations
	batchHash := vm.BatchHash(reqIds, ts, op.peerIndex())
	op.leaderStatus = &leaderStatus{
		reqs:          reqs,
		batchHash:     batchHash,
		timestamp:     ts,
		signedResults: make([]*signedResult, op.committee.Size()),
	}
	op.log.Debugw("runCalculationsAsync leader",
		"batch hash", batchHash.String(),
		"batch", reqIdsStr,
		"ts", ts,
	)
	// process the batch on own (leader) side. Start calculations on VM in a separate thread
	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		leaderPeerIndex: op.committee.OwnPeerIndex(),
		timestamp:       ts,
		accrueFeesTo:    rewardAddress,
	})
	// the LeaderCalculationsStarted stage means at least a quorum of async
	// calculation tasks has been started: locally and on peers
	op.setNextConsensusStage(consensusStageLeaderCalculationsStarted)
}

// checkQuorum takes an action if quorum of results and partial signatures has been reached.
// If so, it aggregates all signatures and produces final transaction.
// The transaction is posted to goshimmer and peers are notified about the fact.
// Note that posting does not mean the transactions reached the goshimmer and/or was started processed
// by the network
func (op *operator) checkQuorum() {
	if op.consensusStage != consensusStageLeaderCalculationsFinished {
		// checking quorum only if leader calculations has been finished
		return
	}
	if op.leaderStatus == nil || op.leaderStatus.resultTxEssence == nil || op.leaderStatus.finalized {
		return
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
			op.log.Warnf("WRONG EssenceHash from peer #%d: %s. Expected: %s",
				i, op.leaderStatus.signedResults[i].essenceHash.String(), mainHash.String())
			op.leaderStatus.signedResults[i] = nil // ignoring
			continue
		}
		err := op.committee.DKShare().VerifySigShare(op.leaderStatus.resultTxEssence.Bytes(), op.leaderStatus.signedResults[i].sigShare)
		if err != nil {
			// TODO here we are ignoring wrong signatures. In general, it means it is an attack
			// In the future when each message will be signed by the peer's identity, the invalidity
			// of the BLS signature means the node is misbehaving.
			op.log.Warnf("WRONG signature from peer #%d: %v", i, err)
			op.leaderStatus.signedResults[i] = nil // ignoring
			continue
		}
		sigShare := op.leaderStatus.signedResults[i].sigShare
		idx, _ := sigShare.Index()
		sigShares = append(sigShares, sigShare)
		contributingPeers = append(contributingPeers, uint16(idx))
	}

	if len(sigShares) < int(op.quorum()) {
		// the quorum has not been reached yet
		return
	}
	// quorum detected

	// finalizing result transaction with signatures
	finalTx, err := op.aggregateSigShares(sigShares)
	if err != nil {
		// should not normally happen
		op.log.Errorf("aggregateSigShares returned: %v", err)
		return
	}

	// check consistency and log
	chainOut, err := utxoutil.GetSingleChainedAliasOutput(finalTx)
	if err != nil {
		op.log.Panic(xerrors.Errorf("major inconsistency: %w", err))
	}
	resultingStateHash, err := hashing.HashValueFromBytes(chainOut.GetStateData())
	if err != nil {
		op.log.Panic(xerrors.Errorf("major inconsistency: %w", err))
	}

	op.log.Infow("FINALIZED RESULT",
		"txid:", finalTx.ID().Base58(),
		"state index", chainOut.GetStateIndex(),
		"state hash", resultingStateHash.String(),
		"contributors", contributingPeers,
	)

	op.leaderStatus.finalized = true

	// posting finalized transaction to goshimmer
	if len(finalTx.Bytes()) > parameters.MaxSerializedTransactionToGoshimmer {
		op.log.Warnf("transaction too large")
		return
	}

	op.nodeConn.PostTransaction(finalTx, op.chain.ID().AsAddress(), op.committee.OwnPeerIndex())
	op.log.Debugf("result transaction has been posted to node. txid: %s", finalTx.ID().Base58())

	// notify peers about finalization of the transaction
	msgData := util.MustBytes(&chain.NotifyFinalResultPostedMsg{
		StateOutputID: op.stateOutput.ID(),
		TxId:          finalTx.ID(),
	})

	numSent := op.committee.SendMsgToPeers(chain.MsgNotifyFinalResultPosted, msgData, time.Now().UnixNano())
	op.log.Debugf("%d peers has been notified about finalized result", numSent)

	op.setNextConsensusStage(consensusStageLeaderResultFinalized)
	op.setFinalizedTransaction(finalTx.ID())

	return
}

// sets new currentState transaction and initializes respective variables
func (op *operator) setNewSCState(msg *chain.StateTransitionMsg) {
	op.stateOutput = msg.ChainOutput
	op.stateTimestamp = msg.Timestamp
	op.currentState = msg.VariableState
	op.sentResultToLeader = nil
	op.postedResultTxid = nilTxID
	op.resetLeader(op.stateOutput.ID().Bytes())
}

func (op *operator) selectRequestsToProcess() []coretypes.Request {
	preSelection := op.mempool.GetReadyListFull(op.quorum() - 1)
	if len(preSelection) == 0 {
		return nil
	}
	lattice := make([][]bool, len(preSelection))
	for i := range lattice {
		lattice[i] = make([]bool, op.size())
		for peerIndex := range preSelection[i].Seen {
			lattice[i][op.committee.OwnPeerIndex()] = true
			if peerIndex < op.size() {
				lattice[i][peerIndex] = true
			}
		}
	}
	// only first element in preselection
	// it has at least quorum of seen's
	ret := []coretypes.Request{preSelection[0].Request}
	first := lattice[0]
	for reqIdx, req := range preSelection[1:] {
		countIntersect := 0
		for i := range first {
			if first[i] && lattice[reqIdx][i] {
				countIntersect++
			}
		}
		if countIntersect >= int(op.quorum()) {
			ret = append(ret, req.Request)
		}
	}
	op.log.Debugf("requests selected for process: %d out of total ready %d", len(ret), len(preSelection))
	return ret
}
