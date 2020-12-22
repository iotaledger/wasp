// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/packages/vm"
)

// EventStateTransitionMsg is triggered by new currentState transition message sent by currentState manager
func (op *operator) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	op.eventStateTransitionMsgCh <- msg
}
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

func (op *operator) EventBalancesMsg(reqMsg chain.BalancesMsg) {
	op.eventBalancesMsgCh <- reqMsg
}
func (op *operator) eventBalancesMsg(reqMsg chain.BalancesMsg) {
	op.log.Debugf("EventBalancesMsg: balances arrived\n%s", txutil.BalancesToString(reqMsg.Balances))
	if err := op.checkSCToken(reqMsg.Balances); err != nil {
		op.log.Debugf("EventBalancesMsg: balances not included: %v", err)
		return
	}

	op.balances = reqMsg.Balances
	op.requestBalancesDeadline = time.Now().Add(chain.RequestBalancesPeriod)

	op.takeAction()
}

// EventRequestMsg triggered by new request msg from the node
func (op *operator) EventRequestMsg(reqMsg *chain.RequestMsg) {
	op.eventRequestMsgCh <- reqMsg
}
func (op *operator) eventRequestMsg(reqMsg *chain.RequestMsg) {
	op.log.Debugw("EventRequestMsg",
		"reqid", reqMsg.RequestId().Short(),
		"backlog req", len(op.requests),
		"backlog notif", len(op.notificationsBacklog),
	)
	// place request into the backlog list
	req, _ := op.requestFromMsg(reqMsg)
	if req == nil {
		op.log.Warn("received already processed request id = %s", reqMsg.RequestId().Short())
		return
	}
	//if reqMsg.Timelock() != 0 {
	//	req.log.Debugf("TIMELOCKED REQUEST: %s. Nowis (Unix) = %d",
	//		reqMsg.RequestSection().String(reqMsg.RequestID()), time.Now().Unix())
	//}

	op.takeAction()
}

// EventNotifyReqMsg request notification received from the peer
func (op *operator) EventNotifyReqMsg(msg *chain.NotifyReqMsg) {
	op.eventNotifyReqMsgCh <- msg
}
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

// EventStartProcessingBatchMsg leader sent command to start processing the batch
func (op *operator) EventStartProcessingBatchMsg(msg *chain.StartProcessingBatchMsg) {
	op.eventStartProcessingBatchMsgCh <- msg
}
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
		op.log.Warnw("EventStartProcessingBatchMsg: ignored",
			"sender", msg.SenderIndex,
			"state index", stateIndex,
			"iAmTheLeader", true,
			"reqIds", idsShortStr(msg.RequestIds),
		)
		return
	}
	numOrig := len(msg.RequestIds)
	reqs := op.takeFromIds(msg.RequestIds)
	if len(reqs) != numOrig {
		op.log.Warnf("node can't process the batch: some requests are already processed")
		return
	}

	reqs = op.filterNotReadyYet(reqs)
	if len(reqs) != numOrig {
		// do not start processing batch if some of it's requests are not ready yet
		op.log.Warn("node is not ready to process the batch")
		return
	}
	// check timestamp
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

	// start async calculation
	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		timestamp:       msg.Timestamp,
		balances:        msg.Balances,
		accrueFeesTo:    msg.FeeDestination,
		leaderPeerIndex: msg.SenderIndex,
	})
	op.setNextConsensusStage(consensusStageSubCalculationsStarted)
	op.takeAction()
}

// EventResultCalculated VM goroutine finished run and posted this message
// the action is to send the result to the leader
func (op *operator) EventResultCalculated(ctx *chain.VMResultMsg) {
	op.eventResultCalculatedCh <- ctx
}
func (op *operator) eventResultCalculated(ctx *chain.VMResultMsg) {
	op.log.Debugf("eventResultCalculated")

	// check if result belongs to context
	if ctx.Task.ResultBlock.StateIndex() != op.mustStateIndex()+1 {
		// out of context. ignore
		return
	}
	op.log.Debugw("eventResultCalculated",
		"batch size", ctx.Task.ResultBlock.Size(),
		"blockIndex", op.mustStateIndex(),
	)

	// inform state manager about new result batch
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

// EventSignedHashMsg result received from another peer
func (op *operator) EventSignedHashMsg(msg *chain.SignedHashMsg) {
	op.eventSignedHashMsgCh <- msg
}
func (op *operator) eventSignedHashMsg(msg *chain.SignedHashMsg) {
	op.log.Debugw("EventSignedHashMsg",
		"sender", msg.SenderIndex,
		"batch hash", msg.BatchHash.String(),
		"essence hash", msg.EssenceHash.String(),
		"ts", msg.OrigTimestamp,
	)
	if op.leaderStatus == nil {
		op.log.Debugf("EventSignedHashMsg: op.leaderStatus == nil")
		// shouldn't be
		return
	}
	if stateIndex, ok := op.blockIndex(); !ok || msg.BlockIndex != stateIndex {
		// out of context
		op.log.Debugf("EventSignedHashMsg: out of context")
		return
	}
	if msg.BatchHash != op.leaderStatus.batchHash {
		op.log.Errorf("EventSignedHashMsg: msg.BatchHash != op.leaderStatus.batchHash")
		return
	}
	if op.leaderStatus.signedResults[msg.SenderIndex] != nil {
		// repeating message from peer
		op.log.Debugf("EventSignedHashMsg: op.leaderStatus.signedResults[msg.SenderIndex].essenceHash != nil")
		return
	}
	op.leaderStatus.signedResults[msg.SenderIndex] = &signedResult{
		essenceHash: msg.EssenceHash,
		sigShare:    msg.SigShare,
	}
	op.takeAction()
}

// EventNotifyFinalResultPostedMsg is triggered by the message sent by the leader to other peers
// immediately after posting finalized transaction to the tangle.
// The message is used to postpone leader rotation deadline for at least confirmation period
func (op *operator) EventNotifyFinalResultPostedMsg(msg *chain.NotifyFinalResultPostedMsg) {
	op.eventNotifyFinalResultPostedMsgCh <- msg
}
func (op *operator) eventNotifyFinalResultPostedMsg(msg *chain.NotifyFinalResultPostedMsg) {
	op.log.Debugw("EventNotifyFinalResultPostedMsg",
		"sender", msg.SenderIndex,
		"stateIdx", msg.BlockIndex,
		"txid", msg.TxId.String(),
	)
	if stateIndex, ok := op.blockIndex(); !ok || msg.BlockIndex != stateIndex {
		return
	}
	if op.iAmCurrentLeader() {
		// this message is intended to subordinates only
		return
	}
	op.setNextConsensusStage(consensusStageSubResultFinalized)
	op.setFinalizedTransaction(&msg.TxId)
}

func (op *operator) EventTransactionInclusionLevelMsg(msg *chain.TransactionInclusionLevelMsg) {
	op.eventTransactionInclusionLevelMsgCh <- msg
}
func (op *operator) eventTransactionInclusionLevelMsg(msg *chain.TransactionInclusionLevelMsg) {
	op.log.Debugw("EventTransactionInclusionLevelMsg",
		"txid", msg.TxId.String(),
		"level", waspconn.InclusionLevelText(msg.Level),
	)
	op.checkInclusionLevel(msg.TxId, msg.Level)
}

func (op *operator) EventTimerMsg(msg chain.TimerTick) {
	op.eventTimerMsgCh <- msg
}
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
			"timelocked", timelockedToString(op.requestTimelocked()),
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
