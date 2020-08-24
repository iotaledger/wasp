package consensus

import (
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"time"

	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/plugins/publisher"
)

func (op *operator) EventProcessorReady(msg committee.ProcessorIsReady) {
	if op.processorReady {
		return
	}
	progHash, ok := op.getProgramHash()
	op.processorReady = ok && msg.ProgramHash == progHash.String()
	if op.processorReady {
		op.log.Infof("User defined VM processor is ready. Program hash: %s", progHash)
	}
}

// EventStateTransitionMsg is triggered by new currentSCState transition message sent by currentSCState manager
func (op *operator) EventStateTransitionMsg(msg *committee.StateTransitionMsg) {
	op.setNewSCState(msg.StateTransaction, msg.VariableState, msg.Synchronized)

	vh := op.currentSCState.Hash()
	op.log.Infof("STATE FOR CONSENSUS #%d, synced: %v, leader: %d iAmTheLeader: %v tx: %s, state hash: %s, backlog: %d",
		op.mustStateIndex(), msg.Synchronized, op.peerPermutation.Current(), op.iAmCurrentLeader(),
		op.stateTx.ID().String(), vh.String(), len(op.requests))

	// remove all processed requests from the local backlog
	if err := op.deleteCompletedRequests(); err != nil {
		op.log.Errorf("deleteCompletedRequests: %v", err)
		return
	}
	// send backlog to the new leader
	if msg.Synchronized {
		if op.iAmCurrentLeader() {
			op.setConsensusStage(consensusStageLeaderStarting)
		} else {
			op.setConsensusStage(consensusStageSubStarting)
		}
	} else {
		op.setConsensusStage(consensusStageNoSync)
	}
	op.takeAction()

	// check is processor is ready for the current consensusStage. If no, initiate load of the processor
	op.processorReady = false
	progHash, ok := op.getProgramHash()
	if !ok {
		op.log.Warnf("program hash is undefined. Only builtin requests can be processed")
		return
	}
	progHashStr := progHash.String()
	op.processorReady = processor.CheckProcessor(progHashStr)
	if !op.processorReady {
		processor.LoadProcessorAsync(progHashStr, func(err error) {
			if err == nil {
				op.committee.ReceiveMessage(committee.ProcessorIsReady{
					ProgramHash: progHashStr,
				})
				publisher.Publish("vmready", op.committee.Address().String(), progHashStr)
			} else {
				op.log.Warnf("failed to load processor: %v", err)
			}
		})
	}
}

func (op *operator) EventBalancesMsg(reqMsg committee.BalancesMsg) {
	op.log.Debugf("EventBalancesMsg: balances arrived\n%s", util.BalancesToString(reqMsg.Balances))
	op.balances = reqMsg.Balances
	op.requestBalancesDeadline = time.Now().Add(op.committee.Params().RequestBalancesPeriod)

	op.takeAction()
}

// EventRequestMsg triggered by new request msg from the node
func (op *operator) EventRequestMsg(reqMsg *committee.RequestMsg) {
	op.log.Debugw("EventRequestMsg",
		"reqid", reqMsg.RequestId().Short(),
		"backlog req", len(op.requests),
		"backlog notif", len(op.notificationsBacklog),
	)
	// place request into the backlog list
	req, _ := op.requestFromMsg(reqMsg)

	if reqMsg.Timelock() != 0 {
		req.log.Debugf("TIMELOCKED REQUEST: %s. Nowis (Unix) = %d",
			reqMsg.RequestBlock().String(reqMsg.RequestId()), time.Now().Unix())
	}

	op.takeAction()
}

// EventNotifyReqMsg request notification received from the peer
func (op *operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {
	op.log.Debugw("EventNotifyReqMsg",
		"reqIds", idsShortStr(msg.RequestIds),
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	op.storeNotification(msg)
	op.markRequestsNotified([]*committee.NotifyReqMsg{msg})

	op.takeAction()
}

// EventStartProcessingBatchMsg leader sent command to start processing the batch
func (op *operator) EventStartProcessingBatchMsg(msg *committee.StartProcessingBatchMsg) {
	bh := vm.BatchHash(msg.RequestIds, msg.Timestamp, msg.SenderIndex)

	op.log.Debugw("EventStartProcessingBatchMsg",
		"sender", msg.SenderIndex,
		"ts", msg.Timestamp,
		"batch hash", bh.String(),
		"reqIds", idsShortStr(msg.RequestIds),
	)
	stateIndex, ok := op.stateIndex()
	if !ok || msg.StateIndex != stateIndex {
		op.log.Debugf("EventStartProcessingBatchMsg: batch out of context. Won't start processing")
		return
	}
	myLeader, _ := op.currentLeader()
	if msg.SenderIndex != myLeader {
		op.log.Debugf("EventStartProcessingBatchMsg: received from different leader. Won't start processing")
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
	// start async calculation
	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		timestamp:       msg.Timestamp,
		balances:        msg.Balances,
		rewardAddress:   msg.RewardAddress,
		leaderPeerIndex: msg.SenderIndex,
	})
	op.setConsensusStage(consensusStageSubCalculationsStarted)
	op.takeAction()
}

// EventResultCalculated VM goroutine finished run and posted this message
// the action is to send the result to the leader
func (op *operator) EventResultCalculated(ctx *vm.VMTask) {
	op.log.Debugf("eventResultCalculated")

	// check if result belongs to context
	if ctx.ResultBatch.StateIndex() != op.mustStateIndex()+1 {
		// out of context. ignore
		return
	}
	op.log.Debugw("eventResultCalculated",
		"batch size", ctx.ResultBatch.Size(),
		"stateIndex", op.mustStateIndex(),
	)

	// inform state manager about new result batch
	go func() {
		op.committee.ReceiveMessage(committee.PendingBatchMsg{
			Batch: ctx.ResultBatch,
		})
	}()

	// save own result or send to the leader
	if ctx.LeaderPeerIndex == op.committee.OwnPeerIndex() {
		op.saveOwnResult(ctx)
	} else {
		op.sendResultToTheLeader(ctx)
	}
	op.takeAction()
}

// EventSignedHashMsg result received from another peer
func (op *operator) EventSignedHashMsg(msg *committee.SignedHashMsg) {
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
	if stateIndex, ok := op.stateIndex(); !ok || msg.StateIndex != stateIndex {
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
func (op *operator) EventNotifyFinalResultPostedMsg(msg *committee.NotifyFinalResultPostedMsg) {
	op.log.Debugw("EventNotifyFinalResultPostedMsg",
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
		"txid", msg.TxId.String(),
	)
	if stateIndex, ok := op.stateIndex(); !ok || msg.StateIndex != stateIndex {
		return
	}
	if op.iAmCurrentLeader() {
		// this message is intended to subordinates only
		return
	}
	op.setConsensusStage(consensusStageSubResultFinalized)
	op.setFinalizedTransaction(&msg.TxId)
}

func (op *operator) EventTransactionInclusionLevelMsg(msg *committee.TransactionInclusionLevelMsg) {
	op.log.Debugw("EventTransactionInclusionLevelMsg",
		"txid", msg.TxId.String(),
		"level", waspconn.InclusionLevelText(msg.Level),
	)
	op.checkInclusionLevel(msg.TxId, msg.Level)
}

func (op *operator) EventTimerMsg(msg committee.TimerTick) {
	if msg%40 == 0 {
		stateIndex, ok := op.stateIndex()
		si := int32(-1)
		if ok {
			si = int32(stateIndex)
		}
		leader, _ := op.currentLeader()
		op.log.Infow("timer tick",
			"#", msg,
			"state index", si,
			"req backlog", len(op.requests),
			"leader", leader,
			"selection", len(op.selectRequestsToProcess()),
			"notif backlog", len(op.notificationsBacklog),
		)
	}
	if msg%2 == 0 {
		op.takeAction()
	}
}
