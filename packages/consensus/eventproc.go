package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
)

func (op *operator) EventStateTransitionMsg(msg *committee.StateTransitionMsg) {
	if op.variableState != nil {
		if !(op.variableState.StateIndex()+1 == msg.VariableState.StateIndex()) {
			panic("assertion failed: op.variableState.stateIndex()+1 == msg.VariableState.stateIndex()")
		}
	}
	// remove all processed requests from the local backlog
	for _, reqId := range msg.RequestIds {
		op.removeRequest(reqId)
	}
	op.setNewState(msg.StateTransaction, msg.VariableState)

	op.takeAction()
}

func (op *operator) EventBalancesMsg(balances committee.BalancesMsg) {
	op.log.Debugf("EventBalancesMsg")
	op.balances = balances.Balances

	op.takeAction()
}

// triggered by new request msg from the node
func (op *operator) EventRequestMsg(reqMsg *committee.RequestMsg) {
	if err := op.validateRequestBlock(reqMsg); err != nil {
		op.log.Warnw("request block validation failed.Ignored",
			"req", reqMsg.Id().Short(),
			"err", err,
		)
		return
	}
	req := op.requestFromMsg(reqMsg)
	req.log.Debugf("eventRequestMsg: id = %s", reqMsg.Id().Short())

	// include request into own list of the current state
	op.appendRequestIdNotifications(op.committee.OwnPeerIndex(), op.stateTx.MustState().StateIndex(), req.reqId)

	// the current leader is notified about new request
	op.sendRequestNotificationsToLeader([]*request{req})
	op.takeAction()
}

func (op *operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {
	op.log.Debugw("EventNotifyReqMsg",
		"num", len(msg.RequestIds),
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	op.MustValidStateIndex(msg.StateIndex)

	// include all reqids into notifications list
	op.appendRequestIdNotifications(msg.SenderIndex, msg.StateIndex, msg.RequestIds...)
	op.takeAction()
}

func (op *operator) EventStartProcessingReqMsg(msg *committee.StartProcessingReqMsg) {
	op.log.Debugw("EventStartProcessingReqMsg",
		"reqId", msg.RequestId.Short(),
		"sender", msg.SenderIndex,
	)

	op.MustValidStateIndex(msg.StateIndex)

	// run calculations async.
	reqRec, processed := op.requestFromId(msg.RequestId)
	if reqRec.reqMsg == nil || processed {
		return
	}

	// start async calculation
	go op.processRequest(runCalculationsParams{
		req:             reqRec,
		ts:              msg.Timestamp,
		balances:        msg.Balances,
		rewardAddress:   *msg.RewardAddress,
		leaderPeerIndex: msg.SenderIndex,
	})
}

func (op *operator) EventResultCalculated(result *committee.VMOutput) {
	op.log.Debugf("eventResultCalculated")

	ctx := result.Inputs.(*runtimeContext)

	// check if result belongs to context
	if ctx.variableState.StateIndex() != op.stateIndex() {
		// out of context. ignore
		return
	}

	// TODO batch of requests. Now assumed 1 request in the batch
	reqId := ctx.reqMsg[0].Id()
	req, ok := op.requestFromId(reqId)
	if !ok {
		// processed
		return
	}
	ctx.log.Debugw("eventResultCalculated",
		"req", req.reqId.Short(),
		"stateIndex", op.stateIndex(),
	)

	if ctx.leaderPeerIndex == op.committee.OwnPeerIndex() {
		op.saveOwnResult(result)
	} else {
		op.sendResultToTheLeader(result)
	}
	op.takeAction()
}

func (op *operator) EventSignedHashMsg(msg *committee.SignedHashMsg) {
	op.log.Debugw("EventSignedHashMsg",
		"reqId", msg.RequestId.Short(),
		"sender", msg.SenderIndex,
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
	if *msg.RequestId != *op.leaderStatus.req.reqId {
		op.log.Debugf("EventSignedHashMsg: !msg.RequestId.Equal(op.leaderStatus.req.reqId)")
		return
	}
	if !msg.OrigTimestamp.Equal(op.leaderStatus.ts) {
		op.log.Debugw("EventSignedHashMsg: !msg.OrigTimestamp.Equal(op.leaderStatus.ts)",
			"msgTs", msg.OrigTimestamp,
			"ownTs", op.leaderStatus.ts)
		return
	}
	if op.leaderStatus.signedResults[msg.SenderIndex].essenceHash != nil {
		// repeating
		op.log.Debugf("EventSignedHashMsg: op.leaderStatus.signedResults[msg.SenderIndex].essenceHash != nil")
		return
	}
	if req, ok := op.requestFromId(msg.RequestId); ok {
		req.log.Debugw("EventSignedHashMsg",
			"origTS", msg.OrigTimestamp,
			"stateIdx", msg.StateIndex,
		)
	}
	// verify signature share received
	if err := op.dkshare.VerifySigShare(op.leaderStatus.resultTx.EssenceBytes(), msg.SigShare); err != nil {
		op.log.Errorf("wrong signature from peer #%d: %v", msg.SenderIndex, err)
		return
	}
	op.leaderStatus.signedResults[msg.SenderIndex] = signedResult{
		essenceHash: msg.EssenceHash,
		sigShare:    msg.SigShare,
	}
	op.takeAction()
}

func (op *operator) EventTimerMsg(msg committee.TimerTick) {
	if msg%10 == 0 {
		op.takeAction()
	}
}
