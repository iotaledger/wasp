package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"time"
)

const getBalancesTimeout = 1 * time.Second

func (op *operator) takeAction() {
	op.doLeader()
	op.doSubordinate()
}

func (op *operator) doSubordinate() {
	for _, cr := range op.currentStateCompRequests {
		if cr.processed {
			continue
		}
		if cr.req.reqTx == nil {
			continue
		}
		cr.processed = true
		//go op.runCalculationsAsync(cr.reqs, cr.ts, cr.leaderPeerIndex)
	}
}

func (op *operator) doLeader() {
	if op.iAmCurrentLeader() {
		op.startProcessing()
	}
	op.checkQuorum()
}

func (op *operator) startProcessing() {
	if op.balances == nil {
		// shouldn't be
		return
	}
	if op.leaderStatus != nil {
		// request already selected and calculations initialized
		return
	}
	reqs := op.selectRequestsToProcess()
	reqIds := takeIds(reqs)
	if len(reqs) == 0 {
		// can't select request to process
		op.log.Debugf("can't select request to process")
		return
	}
	op.log.Debugw("requests selected to process",
		"stateIdx", op.stateTx.MustState().StateIndex(),
		"batch size", len(reqs),
	)
	msgData := hashing.MustBytes(&committee.StartProcessingReqMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			// ts is set by SendMsgToPeers
			StateIndex: op.stateTx.MustState().StateIndex(),
		},
		RewardAddress: *registry.GetRewardAddress(op.committee.Address()),
		Balances:      op.balances,
		RequestIds:    reqIds,
	})

	numSucc, ts := op.committee.SendMsgToPeers(committee.MsgStartProcessingRequest, msgData)

	op.log.Debugf("%d 'msgStartProcessingRequest' messages sent to peers", numSucc)

	if numSucc < op.quorum()-1 {
		// doesn't make sense to continue because less than quorum sends succeeded
		op.log.Errorf("only %d 'msgStartProcessingRequest' sends succeeded", numSucc)
		return
	}
	batchHash := vm.BatchHash(reqIds, ts)
	op.leaderStatus = &leaderStatus{
		reqs:          reqs,
		batchHash:     batchHash,
		balances:      op.balances,
		timestamp:     ts,
		signedResults: make([]*signedResult, op.committee.Size()),
	}
	op.log.Debugw("runCalculationsAsync leader",
		"batch hash", batchHash.String(),
		"ts", ts,
	)

	op.runCalculationsAsync(runCalculationsParams{
		requests:        reqs,
		leaderPeerIndex: op.committee.OwnPeerIndex(),
		balances:        op.balances,
		timestamp:       ts,
		rewardAddress:   *registry.GetRewardAddress(op.committee.Address()),
	})
}

func (op *operator) checkQuorum() bool {
	op.log.Debug("checkQuorum")
	if op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized {
		//log.Debug("checkQuorum: op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized")
		return false
	}
	// collect signature shares available
	mainHash := op.leaderStatus.signedResults[op.committee.OwnPeerIndex()].essenceHash
	sigShares := make([][]byte, 0, op.committee.Size())
	for i := range op.leaderStatus.signedResults {
		if op.leaderStatus.signedResults[i] == nil {
			continue
		}
		if op.leaderStatus.signedResults[i].essenceHash != mainHash {
			op.log.Warnf("wrong EssenceHash from peer #%d", i)
			op.leaderStatus.signedResults[i] = nil // ignoring
			continue
		}
		err := op.dkshare.VerifySigShare(op.leaderStatus.resultTx.EssenceBytes(), op.leaderStatus.signedResults[i].sigShare)
		if err != nil {
			op.log.Warnf("wrong signature from peer #%d: %v", i, err)
			op.leaderStatus.signedResults[i] = nil // ignoring
			continue
		}

		sigShares = append(sigShares, op.leaderStatus.signedResults[i].sigShare)
	}

	if len(sigShares) < int(op.quorum()) {
		return false
	}
	// quorum detected
	err := op.aggregateSigShares(sigShares)
	if err != nil {
		op.log.Errorf("aggregateSigShares returned: %v", err)
		return false
	}

	if !op.leaderStatus.resultTx.SignaturesValid() {
		op.log.Error("final signature invalid: something went wrong while finalizing result transaction")
		return false
	}

	op.leaderStatus.batch.WithStateTransaction(op.leaderStatus.resultTx.ID())

	addr := op.committee.Address()
	if err := op.leaderStatus.resultTx.ValidateConsumptionOfInputs(&addr, op.leaderStatus.balances); err != nil {
		op.log.Errorf("ValidateConsumptionOfInputs: final tx invalid: %v", err)
		return false
	}

	sh := op.leaderStatus.resultTx.MustState().VariableStateHash()
	op.log.Infof("FINALIZED RESULT. Posting transaction to the Value Tangle. txid = %s state hash = %s",
		op.leaderStatus.resultTx.ID().String(),
		sh.String(),
	)
	op.leaderStatus.finalized = true

	nodeconn.PostTransactionToNodeAsyncWithRetry(op.leaderStatus.resultTx.Transaction, 2*time.Second, 7*time.Second, op.log)
	return true
}

// sets new state transaction and initializes respective variables
func (op *operator) setNewState(stateTx *sctransaction.Transaction, variableState state.VariableState) {
	op.stateTx = stateTx
	op.balances = nil
	op.getBalancesDeadline = time.Now()

	op.variableState = variableState

	op.resetLeader(stateTx.ID().Bytes())

	// if consistently moving to the next state, computation requests and notifications about
	// requests for the next state index are brought to the current state next state list is cleared
	// otherwise any state data is cleared
	if op.variableState != nil && variableState.StateIndex() == op.variableState.StateIndex()+1 {
		op.currentStateCompRequests, op.nextStateCompRequests =
			op.nextStateCompRequests, op.currentStateCompRequests
		op.nextStateCompRequests = op.nextStateCompRequests[:0]
	} else {
		op.nextStateCompRequests = op.nextStateCompRequests[:0]
		op.currentStateCompRequests = op.currentStateCompRequests[:0]
	}

	for _, req := range op.requests {
		setAllFalse(req.notifications)
		req.notifications[op.peerIndex()] = req.reqTx != nil
	}
	// run through the notification backlog and mark relevant notifications
	for _, nmsg := range op.notificationsBacklog {
		if nmsg.StateIndex == op.variableState.StateIndex() {
			for _, rid := range nmsg.RequestIds {
				r, ok := op.requestFromId(*rid)
				if !ok {
					continue
				}
				r.notifications[nmsg.SenderIndex] = true
			}
		}
	}
	// clean notification backlog
	op.notificationsBacklog = op.notificationsBacklog[:0]
}
