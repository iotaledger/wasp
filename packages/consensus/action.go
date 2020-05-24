package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
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
		if cr.req.reqMsg == nil {
			continue
		}
		cr.processed = true
		//go op.processRequest(cr.req, cr.ts, cr.leaderPeerIndex)
	}
}

func (op *operator) doLeader() {
	if op.iAmCurrentLeader() {
		if op.balances == nil {
			// of balances are not known yet, request it from the node
			op.requestBalancesFromNode()
		} else {
			op.startProcessing()
		}
	}
	op.checkQuorum()
}

func (op *operator) requestBalancesFromNode() {
	if op.balances == nil && time.Now().After(op.getBalancesDeadline) {
		addr := op.committee.Address()
		nodeconn.RequestBalancesFromNode(&addr)
		op.getBalancesDeadline = time.Now().Add(getBalancesTimeout)
	}
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
	req := op.selectRequestToProcess()
	if req == nil {
		// can't select request to process
		op.log.Debugf("can't select request to process")
		return
	}
	req.log.Debugw("request selected to process", "stateIdx", op.stateTx.MustState().StateIndex())
	msgData := hashing.MustBytes(&committee.StartProcessingReqMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			// ts is set by SendMsgToPeers
			StateIndex: op.stateTx.MustState().StateIndex(),
		},
		RewardAddress: registry.GetRewardAddress(op.committee.Address()),
		Balances:      op.balances,
		RequestId:     req.reqId,
	})

	numSucc, ts := op.committee.SendMsgToPeers(committee.MsgStartProcessingRequest, msgData)

	req.log.Debugf("%d 'msgStartProcessingRequest' messages sent to peers", numSucc)

	if numSucc < op.quorum()-1 {
		// doesn't make sense to continue because less than quorum sends succeeded
		req.log.Errorf("only %d 'msgStartProcessingRequest' sends succeeded", numSucc)
		return
	}
	op.leaderStatus = &leaderStatus{
		req:           req,
		ts:            ts,
		signedResults: make([]signedResult, op.committee.Size()),
	}
	req.log.Debugf("msgStartProcessingRequest successfully sent to %d peers", numSucc)

	// run calculations async.
	go op.processRequest(runCalculationsParams{
		req:             req,
		ts:              ts,
		balances:        op.balances,
		rewardAddress:   *registry.GetRewardAddress(op.committee.Address()),
		leaderPeerIndex: op.committee.OwnPeerIndex(),
	})
}

func (op *operator) checkQuorum() bool {
	op.log.Debug("checkQuorum")
	if op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized {
		//log.Debug("checkQuorum: op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized")
		return false
	}
	mainHash := op.leaderStatus.signedResults[op.committee.OwnPeerIndex()].essenceHash
	if mainHash == nil {
		//log.Debug("checkQuorum: mainHash == nil")
		return false
	}
	sigShares := make([][]byte, 0, op.committee.Size())
	for i := range op.leaderStatus.signedResults {
		if op.leaderStatus.signedResults[i].essenceHash == nil {
			continue
		}
		if *op.leaderStatus.signedResults[i].essenceHash == *mainHash {
			sigShares = append(sigShares, op.leaderStatus.signedResults[i].sigShare)
		}
	}
	if len(sigShares) < int(op.quorum()) {
		return false
	}
	// quorum detected
	err := op.aggregateSigShares(sigShares)
	if err != nil {
		op.leaderStatus.req.log.Errorf("aggregateSigShares returned: %v", err)
		return false
	}
	if !op.leaderStatus.resultTx.SignaturesValid() {
		op.log.Errorf("something went wrong while finalizing result transaction: %v", err)
		return false
	}

	op.log.Info("FINALIZED RESULT. Posting to the Value Tangle. Req = %s", op.leaderStatus.req.reqId.Short())
	// TODO post to tangle
	return true
}

// sets new state transaction and initializes respective variables
func (op *operator) setNewState(stateTx *sctransaction.Transaction, variableState state.VariableState) {
	op.stateTx = stateTx
	op.balances = nil
	op.getBalancesDeadline = time.Now()

	op.variableState = variableState

	op.resetLeader(stateTx.ID().Bytes())

	// computation requests and notifications about requests for the next state index
	// are brought to the current state next state list is cleared
	op.currentStateCompRequests, op.nextStateCompRequests =
		op.nextStateCompRequests, op.currentStateCompRequests
	op.nextStateCompRequests = op.nextStateCompRequests[:0]

	op.requestNotificationsCurrentState, op.requestNotificationsNextState =
		op.requestNotificationsNextState, op.requestNotificationsCurrentState
	op.requestNotificationsNextState = op.requestNotificationsNextState[:0]
}

func (op *operator) selectRequestToProcess() *request {
	// virtual voting
	votes := make(map[sctransaction.RequestId]int)
	for _, rn := range op.requestNotificationsCurrentState {
		if _, ok := votes[*rn.reqId]; !ok {
			votes[*rn.reqId] = 0
		}
		votes[*rn.reqId] = votes[*rn.reqId] + 1
	}
	if len(votes) == 0 {
		return nil
	}
	maxvotes := 0
	for _, v := range votes {
		if v > maxvotes {
			maxvotes = v
		}
	}
	if maxvotes < int(op.quorum()) {
		return nil
	}
	candidates := make([]*request, 0, len(votes))
	for rid, v := range votes {
		if v == int(op.quorum()) {
			req := op.requests[rid]
			if req.reqMsg != nil {
				candidates = append(candidates, req)
			}
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	sortRequestsByAge(candidates)
	return candidates[0]
}
