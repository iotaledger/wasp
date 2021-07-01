package consensus

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/rotate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"golang.org/x/xerrors"
)

// takeAction triggers actions whenever relevant
func (c *Consensus) takeAction() {
	if !c.workflow.stateReceived || c.workflow.finished {
		c.log.Debugf("takeAction skipped: stateReceived %v, workflow finished %v", c.workflow.stateReceived, c.workflow.finished)
		return
	}
	c.proposeBatchIfNeeded()
	c.runVMIfNeeded()
	c.broadcastSignedResultIfNeeded()
	c.checkQuorum()
	c.postTransactionIfNeeded()
	c.pullInclusionStateIfNeeded()
}

// proposeBatchIfNeeded when non empty ready batch is available is in mempool propose it as a candidate
// for the ACS agreement
func (c *Consensus) proposeBatchIfNeeded() {
	if c.workflow.batchProposalSent {
		c.log.Debugf("proposeBatch not needed: batch proposal already sent")
		return
	}
	if c.workflow.consensusBatchKnown {
		c.log.Debugf("proposeBatch not needed: consensus batch already known")
		return
	}
	if time.Now().Before(c.delayBatchProposalUntil) {
		c.log.Debugf("proposeBatch not needed: delayed till %v", c.delayBatchProposalUntil)
		return
	}
	reqs := c.mempool.ReadyNow()
	if len(reqs) == 0 {
		c.log.Debugf("proposeBatch not needed: no ready requests in mempool")
		return
	}
	c.log.Debugf("proposeBatch needed: ready requests len = %d", len(reqs))
	proposal := c.prepareBatchProposal(reqs)
	// call the ACS consensus. The call should spawn goroutine itself
	c.committee.RunACSConsensus(proposal.Bytes(), c.acsSessionID, c.stateOutput.GetStateIndex(), func(sessionID uint64, acs [][]byte) {
		c.log.Debugf("proposeBatch RunACSConsensus callback: responding to ACS session ID %v: len = %d", sessionID, len(acs))
		go c.chain.ReceiveMessage(&chain.AsynchronousCommonSubsetMsg{
			ProposedBatchesBin: acs,
			SessionID:          sessionID,
		})
	})

	c.log.Infof("proposeBatch: proposed batch len = %d, ACS session ID: %d, state index: %d, proposal: %+v",
		len(reqs), c.acsSessionID, c.stateOutput.GetStateIndex(), coretypes.ShortRequestIDs(proposal.RequestIDs))
	c.workflow.batchProposalSent = true
}

// runVMIfNeeded attempts to extract deterministic batch of requests from ACS.
// If it succeeds (i.e. all requests are available) and the extracted batch is nonempty, it runs the request
func (c *Consensus) runVMIfNeeded() {
	if !c.workflow.consensusBatchKnown {
		c.log.Debugf("runVM not needed: consensus batch is not known")
		return
	}
	if c.workflow.vmStarted || c.workflow.vmResultSigned {
		c.log.Debugf("runVM not needed: vmStarted %v, vmResultSigned %v",
			c.workflow.vmStarted, c.workflow.vmResultSigned)
		return
	}
	if time.Now().Before(c.delayRunVMUntil) {
		c.log.Debugf("runVM not needed: delayed till %v", c.delayRunVMUntil)
		return
	}
	reqs, missingRequestIndexes, allArrived := c.mempool.ReadyFromIDs(c.consensusBatch.Timestamp, c.consensusBatch.RequestIDs...)

	c.missingRequestsMutex.Lock()
	defer c.missingRequestsMutex.Unlock()
	c.missingRequestsFromBatch = make(map[coretypes.RequestID][32]byte) // reset list of missing requests

	if !allArrived {
		// some requests are not ready, so skip VM call this time. Maybe next time will be more luck
		c.delayRunVMUntil = time.Now().Add(c.timers.VMRunRetryToWaitForReadyRequests)
		c.log.Infof("runVM not needed: some requests didn't arrive yet")

		// send message to other committee nodes asking for the missing requests
		if !parameters.GetBool(parameters.PullMissingRequestsFromCommittee) {
			return
		}
		missingRequestIds := []coretypes.RequestID{}
		for _, idx := range missingRequestIndexes {
			reqID := c.consensusBatch.RequestIDs[idx]
			reqHash := c.consensusBatch.RequestHashes[idx]
			c.missingRequestsFromBatch[reqID] = reqHash
			missingRequestIds = append(missingRequestIds, reqID)
		}
		c.log.Debugf("runVMIfNeeded: asking for missing requests, ids: %v", missingRequestIds)
		msgData := chain.NewMissingRequestIDsMsg(&missingRequestIds).Bytes()
		c.committee.SendMsgToPeers(chain.MsgMissingRequestIDs, msgData, time.Now().UnixNano())
	}
	if len(reqs) == 0 {
		// due to change in time, all requests became non processable ACS must be run again
		c.log.Debugf("runVM not needed: empty list of processable requests. Reset workflow")
		c.resetWorkflow()
		return
	}

	c.log.Debugf("runVM needed: total number of requests = %d", len(reqs))
	// here reqs as a set is deterministic. Must be sorted to have fully deterministic list
	sort.Slice(reqs, func(i, j int) bool {
		switch {
		case reqs[i].Nonce() < reqs[j].Nonce():
			return true
		case reqs[i].Nonce() > reqs[j].Nonce():
			return false
		default:
			return bytes.Compare(reqs[i].ID().Bytes(), reqs[j].ID().Bytes()) < 0
		}
	})

	// ensure that no more than 126 on ledger requests are in a batch
	onLedgerCount := 0
	reqsFiltered := make([]coretypes.Request, 0, len(reqs))
	for _, req := range reqs {
		_, isOnLedgerReq := req.(*request.RequestOnLedger)
		if isOnLedgerReq {
			onLedgerCount++
			if onLedgerCount >= ledgerstate.MaxInputCount {
				// do not include more on-ledger requests that number of tx inputs allowed-1 ("-1" for chain input)
				continue
			}
		}
		reqsFiltered = append(reqsFiltered, req)
	}

	c.log.Debugf("runVM: sorted requests and filtered onLedger request overhead, running VM with batch len = %d", len(reqsFiltered))
	if vmTask := c.prepareVMTask(reqsFiltered); vmTask != nil {
		c.workflow.vmStarted = true
		go c.vmRunner.Run(vmTask)
		c.log.Debugf("runVM: VM started")
	} else {
		c.log.Errorf("runVM: error preparing VM task")
	}
}

func (c *Consensus) prepareVMTask(reqs []coretypes.Request) *vm.VMTask {
	task := &vm.VMTask{
		ACSSessionID:       c.acsSessionID,
		Processors:         c.chain.Processors(),
		ChainInput:         c.stateOutput,
		SolidStateBaseline: c.chain.GlobalStateSync().GetSolidIndexBaseline(),
		Entropy:            c.consensusEntropy,
		ValidatorFeeTarget: *c.consensusBatch.FeeDestination,
		Requests:           reqs,
		Timestamp:          c.consensusBatch.Timestamp,
		VirtualState:       c.currentState.Clone(),
		Log:                c.log,
	}
	if !task.SolidStateBaseline.IsValid() {
		c.log.Debugf("prepareVMTask: solid state baseline is invalid. Do not even start the VM")
		return nil
	}
	task.OnFinish = func(_ dict.Dict, err error, vmError error) {
		if vmError != nil {
			c.log.Errorf("runVM OnFinish callback: VM task failed: %v", vmError)
			return
		}
		c.log.Debugf("runVM OnFinish callback: responding by state index: %d state hash: %s",
			task.VirtualState.BlockIndex(), task.VirtualState.Hash())
		c.chain.ReceiveMessage(&chain.VMResultMsg{
			Task: task,
		})
	}
	c.log.Debugf("prepareVMTask: VM task prepared")
	return task
}

func (c *Consensus) broadcastSignedResultIfNeeded() {
	if !c.workflow.vmResultSigned {
		c.log.Debugf("broadcastSignedResult not needed: vm result is not signed")
		return
	}
	if time.Now().After(c.delaySendingSignedResult) {
		signedResult := c.resultSignatures[c.committee.OwnPeerIndex()]
		msg := &chain.SignedResultMsg{
			ChainInputID: c.stateOutput.ID(),
			EssenceHash:  signedResult.EssenceHash,
			SigShare:     signedResult.SigShare,
		}
		c.committee.SendMsgToPeers(chain.MsgSignedResult, util.MustBytes(msg), time.Now().UnixNano())
		c.delaySendingSignedResult = time.Now().Add(c.timers.BroadcastSignedResultRetry)

		c.log.Debugf("broadcastSignedResult: broadcasted: essence hash: %s, chain input %s",
			msg.EssenceHash.String(), coretypes.OID(msg.ChainInputID))
	} else {
		c.log.Debugf("broadcastSignedResult not needed: delayed till %v", c.delaySendingSignedResult)
	}
}

// checkQuorum when relevant check if quorum of signatures to the own calculated result is available
// If so, it aggregates signatures and finalizes the transaction.
// Then it deterministically calculates a priority sequence among contributing nodes for posting
// the transaction to L1. The deadline por posting is set proportionally to the sequence number (deterministic)
// If the node sees the transaction of the L1 before its deadline, it cancels its posting
func (c *Consensus) checkQuorum() {
	if c.workflow.transactionFinalized {
		c.log.Debugf("checkQuorum not needed: transaction already finalized")
		return
	}
	if !c.workflow.vmResultSigned {
		// only can aggregate signatures if own result is calculated
		c.log.Debugf("checkQuorum not needed: vm result is not signed")
		return
	}
	// must be not nil
	ownHash := c.resultSignatures[c.committee.OwnPeerIndex()].EssenceHash
	contributors := make([]uint16, 0, c.committee.Size())
	for i, sig := range c.resultSignatures {
		if sig == nil {
			continue
		}
		if sig.EssenceHash == ownHash {
			contributors = append(contributors, uint16(i))
		} else {
			c.log.Warnf("checkQuorum: wrong essence hash: expected(own): %s, got (from %d): %s", ownHash, i, sig.EssenceHash)
		}
	}
	quorumReached := len(contributors) >= int(c.committee.Quorum())
	c.log.Debugf("checkQuorum for essence hash %v:  contributors %+v, quorum %v reached: %v",
		ownHash.String(), contributors, c.committee.Quorum(), quorumReached)
	if !quorumReached {
		return
	}
	sigSharesToAggregate := make([][]byte, len(contributors))
	invalidSignatures := false
	for i, idx := range contributors {
		err := c.committee.DKShare().VerifySigShare(c.resultTxEssence.Bytes(), c.resultSignatures[idx].SigShare)
		if err != nil {
			// TODO here we are ignoring wrong signatures. In general, it means it is an attack
			//  In the future when each message will be signed by the peer's identity, the invalidity
			//  of the BLS signature means the node is misbehaving and should be punished.
			c.log.Warnf("checkQuorum: INVALID SIGNATURE from peer #%d: %v", i, err)
			invalidSignatures = true
		} else {
			sigSharesToAggregate[i] = c.resultSignatures[idx].SigShare
		}
	}
	if invalidSignatures {
		c.log.Errorf("checkQuorum: some signatures were invalid. Reset workflow")
		c.resetWorkflow()
		return
	}
	c.log.Debugf("checkQuorum: all signatures are valid")
	tx, chainOutput, err := c.finalizeTransaction(sigSharesToAggregate)
	if err != nil {
		c.log.Errorf("checkQuorum finalizeTransaction fail: %v", err)
		return
	}

	c.finalTx = tx

	if !chainOutput.GetIsGovernanceUpdated() {
		// if it is not state controller rotation, sending message to state manager
		// Otherwise state manager is not notified
		chainOutputID := chainOutput.ID()
		go c.chain.ReceiveMessage(&chain.StateCandidateMsg{
			State:             c.resultState,
			ApprovingOutputID: chainOutputID,
		})
		c.log.Debugf("checkQuorum: StateCandidateMsg sent for state index %v, approving output ID %v",
			c.resultState.BlockIndex(), coretypes.OID(chainOutputID))
	}

	// calculate deterministic and pseudo-random order and postTxDeadline among contributors
	var postSeqNumber uint16
	var permutation *util.Permutation16
	if c.iAmContributor {
		permutation = util.NewPermutation16(uint16(len(c.contributors)), tx.ID().Bytes())
		postSeqNumber = permutation.GetArray()[c.myContributionSeqNumber]
		c.postTxDeadline = time.Now().Add(time.Duration(postSeqNumber) * c.timers.PostTxSequenceStep)

		c.log.Debugf("checkQuorum: finalized tx %s, iAmContributor: true, postSeqNum: %d, permutation: %+v",
			tx.ID().Base58(), postSeqNumber, permutation.GetArray())
	} else {
		c.log.Debugf("checkQuorum: finalized tx %s, iAmContributor: false", tx.ID().Base58())
	}
	c.workflow.transactionFinalized = true
	c.pullInclusionStateDeadline = time.Now()
}

// postTransactionIfNeeded posts a finalized transaction upon deadline unless it was evidenced on L1 before the deadline.
func (c *Consensus) postTransactionIfNeeded() {
	if !c.workflow.transactionFinalized {
		c.log.Debugf("postTransaction not needed: transaction is not finalized")
		return
	}
	if !c.iAmContributor {
		// only contributors post transaction
		c.log.Debugf("postTransaction not needed: i am not a contributor")
		return
	}
	if c.workflow.transactionPosted {
		c.log.Debugf("postTransaction not needed: transaction already posted")
		return
	}
	if c.workflow.transactionSeen {
		c.log.Debugf("postTransaction not needed: transaction already seen")
		return
	}
	if time.Now().Before(c.postTxDeadline) {
		c.log.Debugf("postTransaction not needed: delayed till %v", c.postTxDeadline)
		return
	}
	c.nodeConn.PostTransaction(c.finalTx)

	c.workflow.transactionPosted = true
	c.log.Infof("postTransaction: POSTED TRANSACTION: %s, number of inputs: %d, outputs: %d", c.finalTx.ID().Base58(), len(c.finalTx.Essence().Inputs()), len(c.finalTx.Essence().Outputs()))
}

// pullInclusionStateIfNeeded periodic pull to know the inclusions state of the transaction. Note that pulling
// starts immediately after finalization of the transaction, not after posting it
func (c *Consensus) pullInclusionStateIfNeeded() {
	if !c.workflow.transactionFinalized {
		c.log.Debugf("pullInclusionState not needed: transaction is not finalized")
		return
	}
	if c.workflow.transactionSeen {
		c.log.Debugf("pullInclusionState not needed: transaction already seen")
		return
	}
	if time.Now().Before(c.pullInclusionStateDeadline) {
		c.log.Debugf("pullInclusionState not needed: delayed till %v", c.pullInclusionStateDeadline)
		return
	}
	c.nodeConn.PullTransactionInclusionState(c.chain.ID().AsAddress(), c.finalTx.ID())
	c.pullInclusionStateDeadline = time.Now().Add(c.timers.PullInclusionStateRetry)
	c.log.Debugf("pullInclusionState: request for inclusion state sent")
}

// prepareBatchProposal creates a batch proposal structure out of requests
func (c *Consensus) prepareBatchProposal(reqs []coretypes.Request) *BatchProposal {
	ts := time.Now()
	if !ts.After(c.stateTimestamp) {
		ts = c.stateTimestamp.Add(1 * time.Nanosecond)
	}
	consensusManaPledge := identity.ID{}
	accessManaPledge := identity.ID{}
	feeDestination := coretypes.NewAgentID(c.chain.ID().AsAddress(), 0)
	// sign state output ID. It will be used to produce unpredictable entropy in consensus
	sigShare, err := c.committee.DKShare().SignShare(c.stateOutput.ID().Bytes())
	c.assert.RequireNoError(err, fmt.Sprintf("prepareBatchProposal: signing output ID %v failed", coretypes.OID(c.stateOutput.ID())))

	ret := &BatchProposal{
		ValidatorIndex:          c.committee.OwnPeerIndex(),
		StateOutputID:           c.stateOutput.ID(),
		RequestIDs:              make([]coretypes.RequestID, len(reqs)),
		RequestHashes:           make([][32]byte, len(reqs)),
		Timestamp:               ts,
		ConsensusManaPledge:     consensusManaPledge,
		AccessManaPledge:        accessManaPledge,
		FeeDestination:          feeDestination,
		SigShareOfStateOutputID: sigShare,
	}
	for i, req := range reqs {
		ret.RequestIDs[i] = req.ID()
		ret.RequestHashes[i] = req.Hash()
	}

	c.log.Debugf("prepareBatchProposal: proposal prepared")
	return ret
}

// receiveACS processed new ACS received from ACS consensus
//nolint:funlen
func (c *Consensus) receiveACS(values [][]byte, sessionID uint64) {
	if c.acsSessionID != sessionID {
		c.log.Debugf("receiveACS: session id missmatch: expected %v, received %v", c.acsSessionID, sessionID)
		return
	}
	if c.workflow.consensusBatchKnown {
		// should not happen
		c.log.Debugf("receiveACS: consensus batch already known (should not happen)")
		return
	}
	if len(values) < int(c.committee.Quorum()) {
		// should not happen. Something wrong with the ACS layer
		c.log.Errorf("receiveACS: ACS is shorter than required quorum. Ignored")
		c.resetWorkflow()
		return
	}
	// decode ACS
	acs := make([]*BatchProposal, len(values))
	for i, data := range values {
		proposal, err := BatchProposalFromBytes(data)
		if err != nil {
			c.log.Errorf("receiveACS: wrong data received. Whole ACS ignored: %v", err)
			c.resetWorkflow()
			return
		}
		acs[i] = proposal
	}
	contributors := make([]uint16, 0, c.committee.Size())
	contributorSet := make(map[uint16]struct{})
	// validate ACS. Dismiss ACS if inconsistent. Should not happen
	for _, prop := range acs {
		if prop.StateOutputID != c.stateOutput.ID() {
			c.log.Warnf("receiveACS: ACS out of context or consensus failure: expected stateOuptudId: %v, generated stateOutputID: %v ",
				coretypes.OID(c.stateOutput.ID()), coretypes.OID(prop.StateOutputID))
			c.resetWorkflow()
			return
		}
		if prop.ValidatorIndex >= c.committee.Size() {
			c.log.Warnf("receiveACS: wrong validtor index in ACS: committee size is %v, validator index is %v",
				c.committee.Size(), prop.ValidatorIndex)
			c.resetWorkflow()
			return
		}
		contributors = append(contributors, prop.ValidatorIndex)
		if _, already := contributorSet[prop.ValidatorIndex]; already {
			c.log.Errorf("receiveACS: duplicate contributors in ACS")
			c.resetWorkflow()
			return
		}
		contributorSet[prop.ValidatorIndex] = struct{}{}
	}

	// sort contributors for determinism because ACS returns sets ordered randomly
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i] < contributors[j]
	})
	iAmContributor := false
	myContributionSeqNumber := uint16(0)
	for i, contr := range contributors {
		if contr == c.committee.OwnPeerIndex() {
			iAmContributor = true
			myContributionSeqNumber = uint16(i)
		}
	}

	// calculate intersection of proposals
	inBatchSet := calcIntersection(acs, c.committee.Size())
	if len(inBatchSet) == 0 {
		// if intersection is empty, reset workflow and retry after some time. It means not all requests
		// reached nodes and we have give it a time. Should not happen often
		c.log.Warnf("receiveACS: ACS intersection (light) is empty. reset workflow. State index: %d, ACS sessionID %d",
			c.stateOutput.GetStateIndex(), sessionID)
		c.resetWorkflow()
		c.delayBatchProposalUntil = time.Now().Add(c.timers.ProposeBatchRetry)
		return
	}
	// calculate other batch parameters in a deterministic way
	par, err := c.calcBatchParameters(acs)
	if err != nil {
		// should not happen, unless insider attack
		c.log.Errorf("receiveACS: inconsistent ACS. Reset workflow. State index: %d, ACS sessionID %d, reason: %v",
			c.stateOutput.GetStateIndex(), sessionID, err)
		c.resetWorkflow()
		c.delayBatchProposalUntil = time.Now().Add(c.timers.ProposeBatchRetry)
	}
	c.consensusBatch = &BatchProposal{
		ValidatorIndex:      c.committee.OwnPeerIndex(),
		StateOutputID:       c.stateOutput.ID(),
		RequestIDs:          inBatchSet,
		Timestamp:           par.medianTs,
		ConsensusManaPledge: par.consensusPledge,
		AccessManaPledge:    par.accessPledge,
		FeeDestination:      par.feeDestination,
	}
	c.consensusEntropy = par.entropy

	c.iAmContributor = iAmContributor
	c.myContributionSeqNumber = myContributionSeqNumber
	c.contributors = contributors

	c.workflow.consensusBatchKnown = true

	if c.iAmContributor {
		c.log.Debugf("receiveACS: ACS received. Contributors to ACS: %+v, iAmContributor: true, seqnr: %d, reqs: %+v",
			c.contributors, c.myContributionSeqNumber, coretypes.ShortRequestIDs(c.consensusBatch.RequestIDs))
	} else {
		c.log.Debugf("receiveACS: ACS received. Contributors to ACS: %+v, iAmContributor: false, reqs: %+v",
			c.contributors, coretypes.ShortRequestIDs(c.consensusBatch.RequestIDs))
	}

	c.runVMIfNeeded()
}

func (c *Consensus) processInclusionState(msg *chain.InclusionStateMsg) {
	if !c.workflow.transactionFinalized {
		c.log.Debugf("processInclusionState: transaction finalized -> skipping.")
		return
	}
	if msg.TxID != c.finalTx.ID() {
		c.log.Debugf("processInclusionState: current transaction id %v does not match the received one %v -> skipping.",
			c.finalTx.ID().Base58(), msg.TxID.Base58())
		return
	}
	switch msg.State {
	case ledgerstate.Pending:
		c.workflow.transactionSeen = true
		c.log.Debugf("processInclusionState: transaction id %v is pending.", c.finalTx.ID().Base58())
	case ledgerstate.Confirmed:
		c.workflow.transactionSeen = true
		c.workflow.finished = true
		c.refreshConsensusInfo()
		c.log.Debugf("processInclusionState: transaction id %s is confirmed; workflow finished", msg.TxID.Base58())
	case ledgerstate.Rejected:
		c.workflow.transactionSeen = true
		c.log.Infof("processInclusionState: transaction id %s is rejected; restarting consensus.", msg.TxID.Base58())
		c.resetWorkflow()
	}
}

func (c *Consensus) finalizeTransaction(sigSharesToAggregate [][]byte) (*ledgerstate.Transaction, *ledgerstate.AliasOutput, error) {
	signatureWithPK, err := c.committee.DKShare().RecoverFullSignature(sigSharesToAggregate, c.resultTxEssence.Bytes())
	if err != nil {
		return nil, nil, xerrors.Errorf("finalizeTransaction RecoverFullSignature fail: %w", err)
	}
	sigUnlockBlock := ledgerstate.NewSignatureUnlockBlock(ledgerstate.NewBLSSignature(*signatureWithPK))

	// check consistency ---------------- check if chain inputs were consumed
	chainInput := ledgerstate.NewUTXOInput(c.stateOutput.ID())
	indexChainInput := -1
	for i, inp := range c.resultTxEssence.Inputs() {
		if inp.Compare(chainInput) == 0 {
			indexChainInput = i
			break
		}
	}
	c.assert.Require(indexChainInput >= 0, fmt.Sprintf("finalizeTransaction: cannot find tx input for state output %v. major inconsistency", coretypes.OID(c.stateOutput.ID())))
	// check consistency ---------------- end

	blocks := make([]ledgerstate.UnlockBlock, len(c.resultTxEssence.Inputs()))
	for i := range c.resultTxEssence.Inputs() {
		if i == indexChainInput {
			blocks[i] = sigUnlockBlock
		} else {
			blocks[i] = ledgerstate.NewAliasUnlockBlock(uint16(indexChainInput))
		}
	}
	tx := ledgerstate.NewTransaction(c.resultTxEssence, blocks)
	chained := transaction.GetAliasOutput(tx, c.chain.ID().AsAddress())
	c.log.Debugf("finalizeTransaction: transaction %v finalized; approving output ID: %v", tx.ID().Base58(), coretypes.OID(chained.ID()))
	return tx, chained, nil
}

func (c *Consensus) setNewState(msg *chain.StateTransitionMsg) {
	c.stateOutput = msg.StateOutput
	c.currentState = msg.State
	c.stateTimestamp = msg.StateTimestamp
	c.acsSessionID = util.MustUint64From8Bytes(hashing.HashData(msg.StateOutput.ID().Bytes()).Bytes()[:8])
	r := ""
	if c.stateOutput.GetIsGovernanceUpdated() {
		r = " (rotate) "
	}
	c.log.Debugf("SET NEW STATE #%d%s, output: %s, hash: %s",
		msg.StateOutput.GetStateIndex(), r, coretypes.OID(msg.StateOutput.ID()), msg.State.Hash().String())
	c.resetWorkflow()
}

func (c *Consensus) resetWorkflow() {
	for i := range c.resultSignatures {
		c.resultSignatures[i] = nil
	}
	c.acsSessionID++
	c.resultState = nil
	c.resultTxEssence = nil
	c.finalTx = nil
	c.consensusBatch = nil
	c.contributors = nil
	c.workflow = workflowFlags{
		stateReceived: c.stateOutput != nil,
	}
	c.log.Debugf("Workflow reset")
}

func (c *Consensus) processVMResult(result *vm.VMTask) {
	if !c.workflow.vmStarted ||
		c.workflow.vmResultSigned ||
		c.acsSessionID != result.ACSSessionID {
		// out of context
		c.log.Debugf("processVMResult: out of context vmStarted %v, vmResultSignedAndBroadcasted %v, expected ACS session ID %v, returned ACS session ID %v",
			c.workflow.vmStarted, c.workflow.vmResultSigned, c.acsSessionID, result.ACSSessionID)
		return
	}
	rotation := result.RotationAddress != nil
	if rotation {
		// if VM returned rotation, we ignore the updated virtual state and produce governance state controller
		// rotation transaction. It does not change state
		c.resultTxEssence = c.makeRotateStateControllerTransaction(result)
		c.resultState = nil
	} else {
		// It is and ordinary state transition
		c.assert.Require(result.ResultTransactionEssence != nil, "processVMResult: result.ResultTransactionEssence != nil")
		c.resultTxEssence = result.ResultTransactionEssence
		c.resultState = result.VirtualState
	}

	essenceBytes := c.resultTxEssence.Bytes()
	essenceHash := hashing.HashData(essenceBytes)
	c.log.Debugf("processVMResult: essence hash: %s. rotate state controller: %v", essenceHash, rotation)

	sigShare, err := c.committee.DKShare().SignShare(essenceBytes)
	c.assert.RequireNoError(err, "processVMResult: ")

	c.resultSignatures[c.committee.OwnPeerIndex()] = &chain.SignedResultMsg{
		SenderIndex:  c.committee.OwnPeerIndex(),
		ChainInputID: result.ChainInput.ID(),
		EssenceHash:  essenceHash,
		SigShare:     sigShare,
	}

	c.workflow.vmResultSigned = true

	c.log.Debugf("processVMResult signed: essence hash: %s", essenceHash.String())
}

func (c *Consensus) makeRotateStateControllerTransaction(task *vm.VMTask) *ledgerstate.TransactionEssence {
	c.log.Debugf("makeRotateStateControllerTransaction: %s", task.RotationAddress.Base58())

	// TODO access and consensus pledge
	essence, err := rotate.MakeRotateStateControllerTransaction(
		task.RotationAddress,
		task.ChainInput,
		task.Timestamp,
		identity.ID{},
		identity.ID{},
	)
	c.assert.RequireNoError(err, "makeRotateStateControllerTransaction: ")
	return essence
}

func (c *Consensus) receiveSignedResult(msg *chain.SignedResultMsg) {
	if c.resultSignatures[msg.SenderIndex] != nil {
		if c.resultSignatures[msg.SenderIndex].EssenceHash != msg.EssenceHash ||
			!bytes.Equal(c.resultSignatures[msg.SenderIndex].SigShare[:], msg.SigShare[:]) {
			c.log.Errorf("receiveSignedResult: conflicting signed result from peer #%d", msg.SenderIndex)
		} else {
			c.log.Warnf("receiveSignedResult: duplicated signed result from peer #%d", msg.SenderIndex)
		}
		return
	}
	if c.stateOutput == nil {
		c.log.Warnf("receiveSignedResult: chain input ID %v received, but state output is nil",
			coretypes.OID(msg.ChainInputID))
		return
	}
	if msg.ChainInputID != c.stateOutput.ID() {
		c.log.Warnf("receiveSignedResult: wrong chain input ID: expected %v, received %v",
			coretypes.OID(c.stateOutput.ID()), coretypes.OID(msg.ChainInputID))
		return
	}
	idx, err := msg.SigShare.Index()
	if err != nil ||
		uint16(idx) >= c.committee.Size() ||
		uint16(idx) == c.committee.OwnPeerIndex() ||
		uint16(idx) != msg.SenderIndex {
		c.log.Errorf("receiveSignedResult: wrong sig share from peer #%d", msg.SenderIndex)
		return
	}
	c.resultSignatures[msg.SenderIndex] = msg
	c.log.Debugf("receiveSignedResult: stored sig share from sender %d, essenceHash %v", msg.SenderIndex, msg.EssenceHash)
}

// ShouldReceiveMissingRequest returns whether or not a request is missing, if the incoming request matches the expetec ID/Hash it is removed from the list
func (c *Consensus) ShouldReceiveMissingRequest(req coretypes.Request) bool {
	c.missingRequestsMutex.Lock()
	defer c.missingRequestsMutex.Unlock()
	expectedHash, exists := c.missingRequestsFromBatch[req.ID()]
	if !exists {
		return false
	}
	reqHash := req.Hash()
	result := bytes.Equal(expectedHash[:], reqHash[:])
	if result {
		delete(c.missingRequestsFromBatch, req.ID())
	}
	return result
}
