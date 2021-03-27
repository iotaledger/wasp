// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/txstream"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/packages/vm"
)

// EventStateTransitionMsg is called when new state transition message sent by the state manager
func (op *operator) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	op.eventStateTransitionMsgCh <- msg
}

// eventStateTransitionMsg internal event handler
func (op *operator) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	op.setNewSCState(msg.AnchorTransaction, msg.VariableState, msg.Synchronized)

	vh := op.currentState.Hash()
	op.log.Infof("STATE FOR CONSENSUS #%d, synced: %v, leader: %d iAmTheLeader: %v tx: %s, state hash: %s, backlog: %d",
		op.mustStateIndex(), msg.Synchronized, op.peerPermutation.Current(), op.iAmCurrentLeader(),
		op.stateTx.ID().String(), vh.String(), len(op.requests))

	// remove all processed requests from the local backlog
	if err := op.deleteCompletedRequests(); err != nil {
		op.log.Errorf("deleteCompletedRequests: %v", err)
		return
	}
	if msg.Synchronized {
		if op.iAmCurrentLeader() {
			op.setNextConsensusStage(consensusStageLeaderStarting)
		} else {
			op.setNextConsensusStage(consensusStageSubStarting)
		}
	} else {
		op.setNextConsensusStage(consensusStageNoSync)
	}
	op.takeAction()
}

// EventBalancesMsg is triggered whenever address balances are coming from the goshimmer
func (op *operator) EventBalancesMsg(reqMsg chain.BalancesMsg) {
	op.eventBalancesMsgCh <- reqMsg
}

// eventBalancesMsg internal event handler
func (op *operator) eventBalancesMsg(reqMsg chain.BalancesMsg) {
	op.log.Debugf("EventBalancesMsg: balances arrived\n%s", txutil.BalancesToString(reqMsg.Balances))

	// TODO here redundant. Should be checked in the dispatcher by tx.Properties (?)
	//if err := op.checkChainToken(reqMsg.Inputs); err != nil {
	//	op.log.Debugf("EventBalancesMsg: balances not included: %v", err)
	//	return
	//}
	op.balances = reqMsg.Balances
	op.requestBalancesDeadline = time.Now().Add(chain.RequestBalancesPeriod)
	op.takeAction()
}

// EventRequestMsg triggered by new request msg from the node
func (op *operator) EventRequestMsg(reqMsg *chain.RequestMsg) {
	op.eventRequestMsgCh <- reqMsg
}

// eventRequestMsg internal handler
func (op *operator) eventRequestMsg(reqMsg *chain.RequestMsg) {
	op.log.Debugw("EventRequestMsg",
		"reqid", reqMsg.RequestId().Short(),
		"backlog req", len(op.requests),
		"backlog notif", len(op.notificationsBacklog),
		"free tokens attached", reqMsg.FreeTokens != nil,
	)
	// place request into the backlog
	req, _ := op.requestFromMsg(reqMsg)
	if req == nil {
		op.log.Warn("received already processed request id = %s", reqMsg.RequestId().Short())
		return
	}
	op.takeAction()
}

// EventNotifyReqMsg request notification received from the peer
func (op *operator) EventNotifyReqMsg(msg *chain.NotifyReqMsg) {
	op.eventNotifyReqMsgCh <- msg
}

// eventNotifyReqMsg internal handler
func (op *operator) eventNotifyReqMsg(msg *chain.NotifyReqMsg) {
	op.log.Debugw("EventNotifyReqMsg",
		"reqIds", idsShortStr(msg.RequestIDs),
		"sender", msg.SenderIndex,
		"stateIdx", msg.BlockIndex,
	)
	op.storeNotification(msg)
	op.markRequestsNotified([]*chain.NotifyReqMsg{msg})
	op.takeAction()
}

// EventStartProcessingBatchMsg command to start processing the batch received from the leader
func (op *operator) EventStartProcessingBatchMsg(msg *chain.StartProcessingBatchMsg) {
	op.eventStartProcessingBatchMsgCh <- msg
}

// eventStartProcessingBatchMsg internal handler
func (op *operator) eventStartProcessingBatchMsg(msg *chain.StartProcessingBatchMsg) {
	bh := vm.BatchHash(msg.RequestIds, msg.Timestamp, msg.SenderIndex)

	op.log.Debugw("EventStartProcessingBatchMsg",
		"sender", msg.SenderIndex,
		"ts", msg.Timestamp,
		"batch hash", bh.String(),
		"reqIds", idsShortStr(msg.RequestIds),
	)
	stateIndex, ok := op.blockIndex()
	if !ok || msg.BlockIndex != stateIndex {
		op.log.Debugf("EventStartProcessingBatchMsg: batch out of context. Won't start processing")
		return
	}
	if op.iAmCurrentLeader() {
		// TODO should not happen. Probably redundant. panic?
		op.log.Warnw("EventStartProcessingBatchMsg: ignored",
			"sender", msg.SenderIndex,
			"state index", stateIndex,
			"iAmTheLeader", true,
			"reqIds", idsShortStr(msg.RequestIds),
		)
		return
	}
	numOrig := len(msg.RequestIds)
	reqs := op.collectProcessableBatch(msg.RequestIds)
	if len(reqs) != numOrig {
		// some request were filtered out because not messages didn't reach the node yet.
		// can't happen? Redundant? panic?
		op.log.Warnf("node can't process the batch: some requests are not known to the node")
		return
	}
	// TODO remove
	//reqs = op.filterNotReadyYet(reqs)
	//if len(reqs) != numOrig {
	//	// do not start processing batch if some of it's requests are not ready yet
	//	op.log.Warn("node is not ready to process the batch")
	//	return
	//}

	// check timestamp. If the local clock is different from the timestamp from the leader more
	// tha threshold, ignore command from the leader.
	// Note that if leader's clock is ot synced with the peers clock significantly, committee
	// will ignore the leader and leader will never earn reward.
	// TODO: attack analysis
	localts := time.Now().UnixNano()
	diff := localts - msg.Timestamp
	if diff < 0 {
		diff = -diff
	}
	if diff > chain.MaxClockDifferenceAllowed.Nanoseconds() {
		op.log.Warn("reject consensus on timestamp: clock difference is too big. Leader ts: %d, local ts: %d, diff: %d",
			msg.Timestamp, localts, diff)
		return
	}

	// start async calculation as requested by the leader
	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		timestamp:       msg.Timestamp,
		balances:        msg.Inputs,
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
	if ctx.Task.ResultBlock.StateIndex() != op.mustStateIndex()+1 {
		// out of context. ignore
		return
	}
	op.log.Debugw("eventResultCalculated",
		"batch size", ctx.Task.ResultBlock.Size(),
		"blockIndex", op.mustStateIndex(),
	)

	// inform own state manager about new result block. The state manager will start waiting
	// from confirmation of it from the tangle
	go func() {
		op.chain.ReceiveMessage(chain.PendingBlockMsg{
			Block: ctx.Task.ResultBlock,
		})
	}()

	// save own result or send to the leader
	if ctx.Leader == op.chain.OwnPeerIndex() {
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
	if stateIndex, ok := op.blockIndex(); !ok || msg.BlockIndex != stateIndex {
		// out of context, current node is lagging behind
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
		"stateIdx", msg.BlockIndex,
		"txid", msg.TxId.String(),
	)
	if stateIndex, ok := op.blockIndex(); !ok || msg.BlockIndex != stateIndex {
		// the leader probably is lagging behind or it is an attack
		return
	}
	if op.iAmCurrentLeader() {
		// this message is intended to subordinates only
		return
	}
	op.setNextConsensusStage(consensusStageSubResultFinalized)
	op.setFinalizedTransaction(&msg.TxId)
}

// EventTransactionInclusionLevelMsg goshimmer send information about transaction
func (op *operator) EventTransactionInclusionLevelMsg(msg *txstream.MsgTxInclusionState) {
	op.eventTransactionInclusionLevelMsgCh <- msg
}

// eventTransactionInclusionLevelMsg intrenal handler
func (op *operator) eventTransactionInclusionLevelMsg(msg *txstream.MsgTxInclusionState) {
	op.log.Debugw("EventTransactionInclusionLevelMsg",
		"txid", msg.TxId.String(),
		"level", waspconn.InclusionLevelText(msg.Level),
	)
	op.checkInclusionLevel(msg.TxId, msg.Level)
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
			"req backlog", len(op.requests),
			"leader", leader,
			"selection", len(op.selectRequestsToProcess()),
			"timelocked", timelockedToString(op.requestsTimeLocked()),
			"notif backlog", len(op.notificationsBacklog),
		)
	}
	if msg%2 == 0 {
		op.takeAction()
	}
}

func timelockedToString(reqs []*request) string {
	if len(reqs) == 0 {
		return "[]"
	}
	ret := make([]string, len(reqs))
	nowis := uint32(time.Now().Unix())
	for i := range ret {
		ret[i] = fmt.Sprintf("%s: %d (-%d)", reqs[i].reqId.Short(), reqs[i].timelock(), reqs[i].timelock()-nowis)
	}
	return fmt.Sprintf("now: %d, [%s]", nowis, strings.Join(ret, ","))
}
