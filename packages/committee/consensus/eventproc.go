package consensus

import (
	"github.com/iotaledger/wasp/packages/hashing"
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

// EventStateTransitionMsg is triggered by new currentState transition message sent by currentState manager
func (op *operator) EventStateTransitionMsg(msg *committee.StateTransitionMsg) {
	op.setNewState(msg.StateTransaction, msg.VariableState, msg.Synchronized)

	vh := op.currentState.Hash()
	op.log.Infof("STATE FOR CONSENSUS #%d, synced: %v, leader: %d iAmTheLeader: %v",
		op.mustStateIndex(), msg.Synchronized, op.peerPermutation.Current(), op.iAmCurrentLeader())
	op.log.Debugf("STATE FOR CONSENSUS #%d, currentState txid: %s, currentState hash: %s",
		op.mustStateIndex(), op.stateTx.ID().String(), vh.String())

	// remove all processed requests from the local backlog
	if err := op.deleteCompletedRequests(); err != nil {
		op.log.Errorf("deleteCompletedRequests: %v", err)
		return
	}
	// send backlog to the new leader
	op.sendNotificationsScheduled = true

	op.takeAction()

	// check is processor is ready for the current state. If no, initiate load of the processor
	op.processorReady = false
	progHash, ok := op.getProgramHash()
	if !ok {
		op.log.Warnf("program hash is undefined. Committee isn't able to load and run VM")
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

	req, _ := op.requestFromMsg(reqMsg)

	if reqMsg.Timelock() != 0 {
		req.log.Debugf("TIMELOCKED REQUEST: %s. Nowis (Unix) = %d",
			reqMsg.RequestBlock().String(reqMsg.RequestId()), time.Now().Unix())
	}

	op.sendNotificationsScheduled = true
	op.takeAction()
}

func (op *operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {
	op.log.Debugw("EventNotifyReqMsg",
		"reqIds", idsShortStr(msg.RequestIds),
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	op.storeNotificationIfNeeded(msg)
	op.markRequestsNotified([]*committee.NotifyReqMsg{msg})

	op.takeAction()
}

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
		op.log.Debugf("EventStartProcessingBatchMsg: batch out of context")
		return
	}

	numOrig := len(msg.RequestIds)
	reqs := op.takeFromIds(msg.RequestIds)
	if len(reqs) != numOrig {
		op.log.Debugf("node can't process the batch: some requests are already processed")
		return
	}
	reqs = op.filterNotReadyYet(reqs)
	if len(reqs) != numOrig {
		op.log.Debugf("node is not ready to process the batch")
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
}

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

	// inform currentState manager about new result batch
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
		return
	}
	if msg.BatchHash != op.leaderStatus.batchHash {
		op.log.Errorf("EventSignedHashMsg: msg.BatchHash != op.leaderStatus.batchHash")
		return
	}
	if op.leaderStatus.signedResults[msg.SenderIndex] != nil {
		// repeating
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
	)
	resTx, ok := op.sentResultsToLeader[msg.SenderIndex]
	if !ok {
		// this is controversial: shall we postpone leader deadline for unseen transaction?
		op.log.Debugf("postpone rotation deadline for unseen transaction for %v more.",
			op.committee.Params().ConfirmationWaitingPeriod)
		op.setLeaderRotationDeadline(op.committee.Params().ConfirmationWaitingPeriod)
		return
	}
	essence := resTx.EssenceBytes()
	if !msg.Signature.IsValid(essence) {
		op.log.Errorf("received invalid final signature from peer #%d. State index: %d, essence hash: %s",
			msg.SenderIndex, msg.StateIndex, hashing.HashData(essence).String())
		return
	}
	op.log.Debugf("valid final signature received: postpone rotation deadline for %v more",
		op.committee.Params().ConfirmationWaitingPeriod)
	op.setLeaderRotationDeadline(op.committee.Params().ConfirmationWaitingPeriod)
}

// EventStateTransactionEvidenced is triggered when state manager receives state transaction not confirmed yet
// It postpones leader rotation deadline
func (op *operator) EventStateTransactionEvidenced(msg *committee.StateTransactionEvidenced) {
	op.log.Debugw("EventStateTransactionEvidenced",
		"txid", msg.TxId.String(),
		"state hash", msg.StateHash.String(),
	)
	if !op.stateTxEvidenced {
		op.stateTxEvidenced = true
		op.setLeaderRotationDeadline(op.committee.Params().ConfirmationWaitingPeriod)
	}
}

func (op *operator) EventTimerMsg(msg committee.TimerTick) {
	if msg%40 == 0 {
		stateIndex, ok := op.stateIndex()
		si := int32(-1)
		if ok {
			si = int32(stateIndex)
		}
		op.log.Infow("timer tick",
			"#", msg,
			"currentState index", si,
			"req backlog", len(op.requests),
			"selection", len(op.selectRequestsToProcess()),
			"notif backlog", len(op.notificationsBacklog),
		)
	}
	if msg%2 == 0 {
		op.takeAction()
	}
}
