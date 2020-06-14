package consensus

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/plugins/publisher"
	"github.com/iotaledger/wasp/plugins/runvm"
)

func (op *operator) EventProcessorReady(msg committee.ProcessorIsReady) {
	op.processorReady = false
	progHashStr, ok := op.getProgramHashStr()
	op.processorReady = ok && msg.ProgramHash == progHashStr
}

// EventStateTransitionMsg is triggered by new state transition message sent by state manager
func (op *operator) EventStateTransitionMsg(msg *committee.StateTransitionMsg) {
	//op.log.Debugf("new varstate:\n%s\n", msg.VariableState.String())

	op.setNewState(msg.StateTransaction, msg.VariableState)

	vh := op.variableState.Hash()
	op.log.Infof("NEW STATE FOR CONSENSUS #%d, leader: %d, state txid: %s, state hash: %s iAmTheLeader: %v",
		op.mustStateIndex(), op.peerPermutation.Current(),
		op.stateTx.ID().String(), vh.String(), op.iAmCurrentLeader())

	// remove all processed requests from the local backlog
	if err := op.deleteCompletedRequests(); err != nil {
		op.log.Errorf("deleteCompletedRequests: %v", err)
		return
	}
	// notify about all request the new leader
	op.sendRequestNotificationsToLeader(nil)

	// check is processor is ready for the current varstate. If no, initiate load of the processor
	op.processorReady = false
	progHashStr, ok := op.getProgramHashStr()
	if !ok {
		op.log.Errorf("major inconsistency: undefined program hash at state #%d", op.mustStateIndex())
		op.committee.Dismiss()
		return
	}

	//op.log.Debugf("++++++++++++++++++++++++++++ proghash = %s", progHashStr)

	op.processorReady = runvm.CheckProcessor(progHashStr)
	if !op.processorReady {
		runvm.LoadProcessorAsync(progHashStr, func(err error) {
			op.committee.ReceiveMessage(committee.ProcessorIsReady{ProgramHash: progHashStr})
			publisher.Publish("loadvm", progHashStr)
		})
	}

	op.takeAction()
}

func (op *operator) EventBalancesMsg(reqMsg committee.BalancesMsg) {
	op.log.Debugf("EventBalancesMsg: balances arrived\n%s", util.BalancesToString(reqMsg.Balances))
	op.balances = reqMsg.Balances

	op.takeAction()
}

// EventRequestMsg triggered by new request msg from the node
func (op *operator) EventRequestMsg(reqMsg committee.RequestMsg) {
	op.log.Debugw("EventRequestMsg", "reqid", reqMsg.RequestId().String())

	if err := op.validateRequestBlock(&reqMsg); err != nil {
		op.log.Warnw("request block validation failed.Ignored",
			"reqs", reqMsg.RequestId().Short(),
			"err", err,
		)
		return
	}
	req, newRequest := op.requestFromMsg(reqMsg)

	if newRequest {
		publisher.Publish("request", "in",
			op.committee.Address().String(),
			reqMsg.Transaction.ID().String(),
			fmt.Sprintf("%d", reqMsg.Index),
		)
	}

	// notify about new request the current leader
	if op.stateTx != nil {
		op.sendRequestNotificationsToLeader([]*request{req})
	}

	op.takeAction()
}

func (op *operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {
	op.log.Debugw("EventNotifyReqMsg",
		"reqIds", idsShortStr(msg.RequestIds),
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	op.notificationsBacklog = append(op.notificationsBacklog, msg)
	op.markRequestsNotified([]*committee.NotifyReqMsg{msg})

	op.takeAction()
}

func (op *operator) EventStartProcessingBatchMsg(msg *committee.StartProcessingBatchMsg) {
	bh := vm.BatchHash(msg.RequestIds, msg.Timestamp)
	op.log.Debugw("EventStartProcessingBatchMsg",
		"sender", msg.SenderIndex,
		"ts", msg.Timestamp,
		"batch hash", bh.String(),
		"reqIds", idsShortStr(msg.RequestIds),
	)
	stateIndex, ok := op.stateIndex()
	if !ok || msg.StateIndex != stateIndex {
		op.log.Debugf("EventStartProcessingBatchMsg: request out of context")
		return
	}

	reqs := make([]*request, len(msg.RequestIds))
	for i := range reqs {
		reqs[i], ok = op.requestFromId(msg.RequestIds[i])
		if !ok {
			op.log.Warn("EventStartProcessingBatchMsg inconsistency: some requests in the batch are already processed")
			return
		}
		if reqs[i].reqTx == nil {
			op.log.Warn("EventStartProcessingBatchMsg inconsistency: some requests in the batch not yet received by the node")
			return
		}
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

func (op *operator) EventTimerMsg(msg committee.TimerTick) {
	if msg%200 == 0 {
		op.log.Debugw("timer tick",
			"#", msg,
			"req backlog", len(op.requests),
			"selection", len(op.selectRequestsToProcess()),
			"notif backlog", len(op.notificationsBacklog),
		)
	}
	if msg%10 == 0 {
		op.takeAction()
	}
}
