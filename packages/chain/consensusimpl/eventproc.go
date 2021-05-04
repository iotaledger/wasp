// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensusimpl

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/vm"
)

// EventStateTransitionMsg is called when new state transition message sent by the state manager
func (op *operator) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	op.eventStateTransitionMsgCh <- msg
}

// eventStateTransitionMsg internal event handler
func (op *operator) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	op.setNewSCState(msg)

	vh := op.currentState.Hash()
	op.log.Infof("STATE FOR CONSENSUS #%d, leader: %d, iAmTheLeader: %v, stateOutput: %s, state hash: %s",
		op.mustStateIndex(), op.peerPermutation.Current(), op.iAmCurrentLeader(),
		coretypes.OID(op.stateOutput.ID()), vh.String())

	if op.iAmCurrentLeader() {
		op.setNextConsensusStage(consensusStageLeaderStarting)
	} else {
		op.setNextConsensusStage(consensusStageSubStarting)
	}
	op.pullBacklogDeadline = time.Now()
	op.takeAction()
}

// EventNotifyReqMsg request notification received from the peer
func (op *operator) EventNotifyReqMsg(msg *chain.NotifyReqMsg) {
	op.eventNotifyReqMsgCh <- msg
}

// eventNotifyReqMsg internal handler
func (op *operator) eventNotifyReqMsg(msg *chain.NotifyReqMsg) {
	op.log.Debugw("EventNotifyReqMsg",
		"reqIds", idsShortStr(msg.RequestIDs...),
		"sender", msg.SenderIndex,
		"stateID", coretypes.OID(msg.StateOutputID),
	)
	if op.stateOutput == nil || op.stateOutput.ID() != msg.StateOutputID {
		op.log.Debugf("EventNotifyReqMsg: out of context")
		return
	}
	for _, rid := range msg.RequestIDs {
		r := rid
		op.mempool.MarkSeenByCommitteePeer(&r, msg.SenderIndex)
	}
	op.takeAction()
}

// EventStartProcessingBatchMsg command to start processing the batch received from the leader
func (op *operator) EventStartProcessingBatchMsg(msg *chain.StartProcessingBatchMsg) {
	op.eventStartProcessingBatchMsgCh <- msg
}

// eventStartProcessingBatchMsg internal handler
func (op *operator) eventStartProcessingBatchMsg(msg *chain.StartProcessingBatchMsg) {
	bh := vm.BatchHash(msg.RequestIDs, time.Unix(0, msg.Timestamp), msg.SenderIndex)

	op.log.Debugw("EventStartProcessingBatchMsg",
		"sender", msg.SenderIndex,
		"ts", msg.Timestamp,
		"batch hash", bh.String(),
		"reqIds", idsShortStr(msg.RequestIDs...),
	)
	if op.stateOutput == nil || op.stateOutput.ID() != msg.StateOutputID {
		op.log.Debugf("EventStartProcessingBatchMsg: batch out of context. Won't start processing")
		return
	}
	if op.iAmCurrentLeader() {
		// TODO should not happen. Probably redundant. panic?
		op.log.Warnw("EventStartProcessingBatchMsg: ignored",
			"sender", msg.SenderIndex,
			"state index", op.stateOutput.GetStateIndex(),
			"iAmTheLeader", true,
			"reqIds", idsShortStr(msg.RequestIDs...),
		)
		return
	}
	reqs, allReady := op.mempool.TakeAllReady(time.Unix(0, msg.Timestamp), msg.RequestIDs...)
	if !allReady {
		op.log.Warnf("node can't process the batch: some requests are not ready on the node")
		return
	}
	// start async calculation as requested by the leader
	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		timestamp:       time.Unix(0, msg.Timestamp),
		accrueFeesTo:    msg.FeeDestination,
		leaderPeerIndex: msg.SenderIndex,
	})
	op.setNextConsensusStage(consensusStageSubCalculationsStarted)
	op.takeAction()
}

// EventResultCalculated batch calculation goroutine finished calculations and posted this message
// the action is to send the result to the leader
func (op *operator) EventResultCalculated(ctx *chain.VMResultMsg) {
	op.eventResultCalculatedCh <- ctx
}

// eventResultCalculated internal handler
func (op *operator) eventResultCalculated(ctx *chain.VMResultMsg) {
	op.log.Debugf("eventResultCalculated")

	// check if result belongs to the context. In general, consensus may be already reached
	// even before the node finished its calculations
	if ctx.Task.VirtualState.BlockIndex() != op.mustStateIndex()+1 {
		// out of context. ignore
		return
	}
	op.log.Debugw("eventResultCalculated",
		"blockIndex", op.mustStateIndex(),
	)
	// inform own state manager about new result block. The state manager will start waiting
	// from confirmation of it from the tangle
	go func() {
		op.chain.ReceiveMessage(chain.StateCandidateMsg{
			State: ctx.Task.VirtualState,
		})
	}()

	// save own result or send to the leader
	if ctx.Leader == op.committee.OwnPeerIndex() {
		op.saveOwnResult(ctx.Task)
	} else {
		op.sendResultToTheLeader(ctx.Task, ctx.Leader)
	}
	op.takeAction()
}

// EventSignedHashMsg partially signed calculation result received from another peer
func (op *operator) EventSignedHashMsg(msg *chain.SignedHashMsg) {
	op.eventSignedHashMsgCh <- msg
}

// eventSignedHashMsg internal handler
func (op *operator) eventSignedHashMsg(msg *chain.SignedHashMsg) {
	op.log.Debugw("EventSignedHashMsg",
		"sender", msg.SenderIndex,
		"batch hash", msg.BatchHash.String(),
		"essence hash", msg.EssenceHash.String(),
		"ts", msg.OrigTimestamp,
	)
	if op.leaderStatus == nil {
		op.log.Debugf("EventSignedHashMsg: op.leaderStatus == nil")
		// shouldn't be, probably an attack
		return
	}
	if op.stateOutput == nil || op.stateOutput.ID() != msg.StateOutputID {
		// notification out of context
		op.log.Debugf("EventSignedHashMsg: out of context")
		return
	}
	if msg.BatchHash != op.leaderStatus.batchHash {
		// shouldn't be. Probably an attack
		op.log.Errorf("EventSignedHashMsg: msg.BatchHash != op.leaderStatus.batchHash")
		return
	}
	if op.leaderStatus.signedResults[msg.SenderIndex] != nil {
		// repeating message from peer.
		// Shouldn't be. May be an attack or misbehavior
		op.log.Debugf("EventSignedHashMsg: op.leaderStatus.signedResults[msg.SenderIndex].essenceHash != nil")
		return
	}
	op.leaderStatus.signedResults[msg.SenderIndex] = &signedResult{
		essenceHash: msg.EssenceHash,
		sigShare:    msg.SigShare,
	}
	op.takeAction()
}

// EventNotifyFinalResultPostedMsg is triggered by the message sent by the leader to peers
// immediately after posting finalized transaction to the tangle.
// The message is used to postpone leader rotation deadline for at least confirmation period
// Note that receiving this message does no mean the transaction has been accepted by th network
func (op *operator) EventNotifyFinalResultPostedMsg(msg *chain.NotifyFinalResultPostedMsg) {
	op.eventNotifyFinalResultPostedMsgCh <- msg
}

// eventNotifyFinalResultPostedMsg internal handler
func (op *operator) eventNotifyFinalResultPostedMsg(msg *chain.NotifyFinalResultPostedMsg) {
	op.log.Debugw("EventNotifyFinalResultPostedMsg",
		"sender", msg.SenderIndex,
		"stateID", coretypes.OID(msg.StateOutputID),
		"txid", msg.TxId.String(),
	)
	if op.stateOutput == nil || op.stateOutput.ID() != msg.StateOutputID {
		// the leader probably is lagging behind or it is an attack
		return
	}
	if op.iAmCurrentLeader() {
		// this message is intended to subordinates only
		return
	}
	op.setNextConsensusStage(consensusStageSubResultFinalized)
	op.setFinalizedTransaction(msg.TxId)
}

// EventTransactionInclusionLevelMsg goshimmer send information about transaction
func (op *operator) EventTransactionInclusionStateMsg(msg *chain.InclusionStateMsg) {
	op.eventTransactionInclusionLevelMsgCh <- msg
}

// eventTransactionInclusionStateMsg internal handler
func (op *operator) eventTransactionInclusionStateMsg(msg *chain.InclusionStateMsg) {
	op.log.Debugw("EventTransactionInclusionStateMsg",
		"txid", msg.TxID.Base58(),
		"level", msg.State.String(),
	)
	op.checkInclusionLevel(&msg.TxID, msg.State)
}

// EventTimerMsg timer tick
func (op *operator) EventTimerMsg(msg chain.TimerTick) {
	op.eventTimerMsgCh <- msg
}

// eventTimerMsg internal handler
func (op *operator) eventTimerMsg(msg chain.TimerTick) {
	if msg%40 == 0 {
		blockIndex, ok := op.blockIndex()
		si := int32(-1)
		if ok {
			si = int32(blockIndex)
		}
		leader, _ := op.currentLeader()
		op.log.Infow("timer tick",
			"#", msg,
			"block index", si,
			"leader", leader,
		)
	}
	if msg%2 == 0 {
		op.takeAction()
	}
}
