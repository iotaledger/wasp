package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/vm"
)

// EventStateTransitionMsg is triggered by new state transition message sent by state manager
func (op *operator) EventStateTransitionMsg(msg *committee.StateTransitionMsg) {
	// remove all processed requests from the local backlog
	for _, reqId := range msg.RequestIds {
		op.removeRequest(reqId)
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

	//op.takeAction()
}

// EventBalancesMsg triggered by balances of the SC address coming from the node
func (op *operator) EventBalancesMsg(balances committee.BalancesMsg) {
	op.log.Debugf("EventBalancesMsg")
	op.balances = balances.Balances

	op.takeAction()
}

// EventRequestMsg triggered by new request msg from the node
func (op *operator) EventRequestMsg(reqMsg *committee.RequestMsg) {
	op.log.Debugw("EventRequestMsg", "reqid", reqMsg.RequestId().String())

	if err := op.validateRequestBlock(reqMsg); err != nil {
		op.log.Warnw("request block validation failed.Ignored",
			"reqs", reqMsg.RequestId().Short(),
			"err", err,
		)
		return
	}
	req := op.requestFromMsg(reqMsg)

	// notify about new request the current leader
	op.sendRequestNotificationsToLeader([]*request{req})

	//op.takeAction()
}

func (op *operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {
	op.log.Debugw("EventNotifyReqMsg",
		"num ids", len(msg.RequestIds),
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	op.MustValidStateIndex(msg.StateIndex)

	op.markRequestsNotified(msg)

	//op.takeAction()
}

func (op *operator) EventStartProcessingReqMsg(msg *committee.StartProcessingReqMsg) {
	op.log.Debugw("EventStartProcessingReqMsg",
		"num reqId", len(msg.RequestIds),
		"sender", msg.SenderIndex,
	)

	op.MustValidStateIndex(msg.StateIndex)

	reqs := make([]*committee.RequestMsg, len(msg.RequestIds))
	for i := range reqs {
		req, ok := op.requestFromId(&msg.RequestIds[i])
		if !ok {
			op.log.Debug("some requests in the batch are already processed")
			return
		}
		if req.reqMsg == nil {
			op.log.Debug("some requests in the batch not yet received by the node")
			return
		}
		reqs = append(reqs, req.reqMsg)
	}
	// start async calculation
	go op.processRequest(runCalculationsParams{
		reqs:            reqs,
		ts:              msg.Timestamp,
		balances:        msg.Balances,
		rewardAddress:   msg.RewardAddress,
		leaderPeerIndex: msg.SenderIndex,
	})
}

func (op *operator) EventResultCalculated(ctx *vm.RuntimeContext) {
	op.log.Debugf("eventResultCalculated")

	// check if result belongs to context
	if ctx.VariableState.StateIndex() != op.stateIndex() {
		// out of context. ignore
		return
	}

	resultBatch, err := state.NewBatch(ctx.StateUpdates, ctx.VariableState.StateIndex()+1)
	if err != nil {
		op.log.Errorf("error while creating batch: %v", err)
		return
	}
	op.log.Debugw("eventResultCalculated",
		"batch size", resultBatch.Size(),
		"stateIndex", op.stateIndex(),
	)

	if ctx.LeaderPeerIndex == op.committee.OwnPeerIndex() {
		op.saveOwnResult(result)
	} else {
		op.sendResultToTheLeader(result)
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
	if !msg.OrigTimestamp.Equal(op.leaderStatus.ts) {
		op.log.Debugw("EventSignedHashMsg: !msg.OrigTimestamp.Equal(op.leaderStatus.ts)",
			"msgTs", msg.OrigTimestamp,
			"ownTs", op.leaderStatus.ts)
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
