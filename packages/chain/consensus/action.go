// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/plugins/nodeconn"
)

// takeAction analyzes the state and updates it and takes action such as sending of message,
// IF IT IS REQUIRED BY THE STATE (for example if deadline achieved, if needed data is not here and similar .
// Is called from timer ticks, also when messages received
func (op *operator) takeAction() {
	op.solidifyRequestArgsIfNeeded()
	op.sendRequestNotificationsToLeader()
	op.startCalculationsAsLeader()
	op.checkQuorum()
	op.rotateLeader()
	op.pullInclusionLevel()
}

// solidifyRequestArgsIfNeeded runs through all requests and, if needed, attempts to solidify args
func (op *operator) solidifyRequestArgsIfNeeded() {
	if time.Now().Before(op.nextArgSolidificationDeadline) {
		return
	}
	reqs := op.allRequests()
	reqs = filterRequests(reqs, func(r *request) bool {
		return r.hasMessage() && !r.hasSolidArgs()
	})
	for _, req := range reqs {
		ok, err := req.reqTx.Requests()[req.reqId.Index()].SolidifyArgs(op.chain.BlobRegistry())
		if err != nil {
			req.log.Errorf("failed to solidify args: %v", err)
		} else {
			req.log.Infof("solidified arguments")
			req.argsSolid = ok
		}
	}
	op.nextArgSolidificationDeadline = time.Now().Add(chain.CheckArgSolidificationEvery)
}

// pullInclusionLevel if it is known that result transaction was posted by the leader,
// some updates from Goshimmer are expected about the status (inclusion level) of the transaction
// If the update about the tx state didn't come as expected (timeout), send the query about it
// to the goshimmer (pull)
func (op *operator) pullInclusionLevel() {
	if op.postedResultTxid == nil {
		return
	}
	if time.Now().After(op.nextPullInclusionLevel) {
		addr := op.chain.Address()
		if err := nodeconn.RequestInclusionLevelFromNode(op.postedResultTxid, &addr); err != nil {
			op.log.Errorf("RequestInclusionLevelFromNode: %v", err)
		}
		op.setNextPullInclusionStageDeadline()
	}
}

// rotateLeader upon expired deadline. The deadline depends on the stage
func (op *operator) rotateLeader() {
	if !op.consensusStageDeadlineExpired() {
		return
	}
	if !op.chain.HasQuorum() {
		op.log.Debugf("leader was not rotated due to no quorum")
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
	if !op.chain.HasQuorum() {
		// no quorum, doesn't make sense to start
		return
	}
	// select requests for the batch
	reqs := op.selectRequestsToProcess()
	if len(reqs) == 0 {
		// empty backlog or nothing is ready
		return
	}
	reqIds := takeIds(reqs)
	reqIdsStr := idsShortStr(reqIds)

	op.log.Debugf("requests selected to process. Current state: %d, Reqs: %+v", op.mustStateIndex(), reqIdsStr)
	rewardAddress := op.getFeeDestination()

	// send to subordinated peers requests to process the batch
	msgData := util.MustBytes(&chain.StartProcessingBatchMsg{
		PeerMsgHeader: chain.PeerMsgHeader{
			// timestamp is set by SendMsgToCommitteePeers
			BlockIndex: op.stateTx.MustState().BlockIndex(),
		},
		FeeDestination: rewardAddress,
		Balances:       op.balances,
		RequestIds:     reqIds,
	})

	// determine timestamp. Must be max(local clock, prev timestamp+1).
	// Adjustment enforced, when needed
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
		balances:      op.balances,
		timestamp:     ts,
		signedResults: make([]*signedResult, op.chain.Size()),
	}
	op.log.Debugw("runCalculationsAsync leader",
		"batch hash", batchHash.String(),
		"batch", reqIdsStr,
		"ts", ts,
	)
	// process the batch on own (leader) side. Start calculations on VM in a separate thread
	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		leaderPeerIndex: op.chain.OwnPeerIndex(),
		balances:        op.balances,
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
	if op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized {
		return
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
			// TODO here we are ignoring wrong signatures. In general, it means it is an attack
			// In the future when each message will be signed by the peer's identity, the invalidity
			// of the BLS signature means the node is misbehaving.
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
		// the quorum has not been reached yet
		return
	}
	// quorum detected

	// finalizing result transaction with signatures
	if err := op.aggregateSigShares(sigShares); err != nil {
		// should not normally happen
		op.log.Errorf("aggregateSigShares returned: %v", err)
		return
	}

	// just in case we are double-checking semantic validity of the transaction
	// Invalidity of properties means internal error
	// Nota that tx ID is not known and cannot be taken before this point,
	_, err := op.leaderStatus.resultTx.Properties()
	if err != nil {
		op.log.Panicf("internal error: invalid tx properties: %v\ndump tx: %s\ndump vtx: %s\n", err,
			op.leaderStatus.resultTx.String(), op.leaderStatus.resultTx.Transaction.String())
		return
	}

	txid := op.leaderStatus.resultTx.ID()
	sh := op.leaderStatus.resultTx.MustState().StateHash()
	stateIndex := op.leaderStatus.resultTx.MustState().BlockIndex()
	op.log.Infof("FINALIZED RESULT. txid: %s, state index: #%d, state hash: %s, contributors: %+v",
		txid.String(), stateIndex, sh.String(), contributingPeers)
	op.leaderStatus.finalized = true

	// posting finalized transaction to goshimmer
	addr := op.chain.Address()
	err = nodeconn.PostTransactionToNode(op.leaderStatus.resultTx.Transaction, &addr, op.chain.OwnPeerIndex())
	if err != nil {
		op.log.Warnf("PostTransactionToNode failed: %v", err)
		return
	}
	op.log.Debugf("result transaction has been posted to node. txid: %s", txid.String())

	// notify peers about finalization of the transaction
	msgData := util.MustBytes(&chain.NotifyFinalResultPostedMsg{
		PeerMsgHeader: chain.PeerMsgHeader{
			// timestamp is set by SendMsgToCommitteePeers
			BlockIndex: op.stateTx.MustState().BlockIndex(),
		},
		TxId: txid,
	})

	numSent := op.chain.SendMsgToCommitteePeers(chain.MsgNotifyFinalResultPosted, msgData, time.Now().UnixNano())
	op.log.Debugf("%d peers has been notified about finalized result", numSent)

	op.setNextConsensusStage(consensusStageLeaderResultFinalized)
	op.setFinalizedTransaction(&txid)

	return
}

// sets new currentState transaction and initializes respective variables
func (op *operator) setNewSCState(stateTx *sctransaction.Transaction, variableState state.VirtualState, synchronized bool) {
	op.stateTx = stateTx
	op.currentState = variableState
	op.sentResultToLeader = nil
	op.postedResultTxid = nil
	op.requestBalancesDeadline = time.Now()
	op.resetLeader(stateTx.ID().Bytes())
	op.adjustNotifications()
}
