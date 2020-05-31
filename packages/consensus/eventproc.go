package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/vm"
)

// EventStateTransitionMsg is triggered by new state transition message sent by state manager
func (op *operator) EventStateTransitionMsg(msg *committee.StateTransitionMsg) {
	// remove all processed requests from the local backlog
	if err := op.deleteCompletedRequests(); err != nil {
		op.log.Errorf("deleteCompletedRequests: %v", err)
		return
	}
	op.setNewState(msg.StateTransaction, msg.VariableState)

	vh := msg.VariableState.Hash()
	op.log.Debugw("EventStateTransitionMsg",
		"state index", msg.VariableState.StateIndex(),
		"state hash", vh.String(),
		"state txid", msg.StateTransaction.ID().String(),
		"num reqs", len(msg.RequestIds),
		"iAmTheLeader", op.iAmCurrentLeader(),
	)
	// notify about all request the new leader
	op.sendRequestNotificationsToLeader(nil)

	op.takeAction()
}

func (op *operator) EventBalancesMsg(reqMsg committee.BalancesMsg) {
	op.log.Debug("EventBalancesMsg")

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
	req := op.requestFromMsg(reqMsg)

	// notify about new request the current leader
	if op.stateTx != nil {
		op.sendRequestNotificationsToLeader([]*request{req})
	}

	op.takeAction()
}

func (op *operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {
	op.log.Debugw("EventNotifyReqMsg",
		"num ids", len(msg.RequestIds),
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	op.MustValidStateIndex(msg.StateIndex)

	op.markRequestsNotified(msg)

	op.takeAction()
}

func (op *operator) EventStartProcessingReqMsg(msg *committee.StartProcessingReqMsg) {
	op.log.Debugw("EventStartProcessingReqMsg",
		"num reqId", len(msg.RequestIds),
		"sender", msg.SenderIndex,
	)

	op.MustValidStateIndex(msg.StateIndex)

	reqs := make([]*request, len(msg.RequestIds))
	for i := range reqs {
		req, ok := op.requestFromId(msg.RequestIds[i])
		if !ok {
			op.log.Debug("some requests in the batch are already processed")
			return
		}
		if req.reqTx == nil {
			op.log.Debug("some requests in the batch not yet received by the node")
			return
		}
		reqs = append(reqs, req)
	}
	// start async calculation
	go op.runCalculationsAsync(runCalculationsParams{
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
	if ctx.ResultBatch.StateIndex() != op.stateIndex()+1 {
		// out of context. ignore
		return
	}
	op.log.Debugw("eventResultCalculated",
		"batch size", ctx.ResultBatch.Size(),
		"stateIndex", op.stateIndex(),
	)
	// TODO what if len(ctx.ResultBatch.Size() == 0 ?

	// save own result
	// or send to the leader
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
	)
	if op.leaderStatus == nil {
		op.log.Debugf("EventSignedHashMsg: op.leaderStatus == nil")
		// shouldn't be
		return
	}
	if msg.StateIndex != op.stateIndex() {
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
	// verify signature share received
	if err := op.dkshare.VerifySigShare(op.leaderStatus.resultTx.EssenceBytes(), msg.SigShare); err != nil {
		op.log.Errorf("wrong signature from peer #%d: %v", msg.SenderIndex, err)
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
		//op.takeAction()
	}
}
