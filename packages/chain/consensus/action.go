// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/identity"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/rotate"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// takeAction triggers actions whenever relevant
func (c *consensus) takeAction() {
	if !c.workflow.IsStateReceived() || !c.workflow.IsInProgress() {
		c.log.Debugf("takeAction skipped: stateReceived: %v, workflow in progress: %v",
			c.workflow.IsStateReceived(), c.workflow.IsInProgress())
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
func (c *consensus) proposeBatchIfNeeded() {
	if c.workflow.IsBatchProposalSent() {
		c.log.Debugf("proposeBatch not needed: batch proposal already sent")
		return
	}
	if c.workflow.IsConsensusBatchKnown() {
		c.log.Debugf("proposeBatch not needed: consensus batch already known")
		return
	}
	if time.Now().Before(c.delayBatchProposalUntil) {
		c.log.Debugf("proposeBatch not needed: delayed till %v", c.delayBatchProposalUntil)
		return
	}
	if time.Now().Before(c.stateTimestamp.Add(c.timers.ProposeBatchDelayForNewState)) {
		c.log.Debugf("proposeBatch not needed: delayed for %v from %v", c.timers.ProposeBatchDelayForNewState, c.stateTimestamp)
		return
	}
	if c.timeData == nil {
		c.log.Debugf("proposeBatch not needed: time data hasn't been received yet")
		return
	}
	reqs := c.mempool.ReadyNow(*c.timeData)
	if len(reqs) == 0 {
		c.log.Debugf("proposeBatch not needed: no ready requests in mempool")
		return
	}
	c.log.Debugf("proposeBatch needed: ready requests len = %d", len(reqs))
	proposal := c.prepareBatchProposal(reqs)
	// call the ACS consensus. The call should spawn goroutine itself
	c.committee.RunACSConsensus(proposal.Bytes(), c.acsSessionID, c.stateOutput.GetStateIndex(), func(sessionID uint64, acs [][]byte) {
		c.log.Debugf("proposeBatch RunACSConsensus callback: responding to ACS session ID %v: len = %d", sessionID, len(acs))
		go c.EnqueueAsynchronousCommonSubsetMsg(&messages.AsynchronousCommonSubsetMsg{
			ProposedBatchesBin: acs,
			SessionID:          sessionID,
		})
	})

	c.log.Infof("proposeBatch: proposed batch len = %d, ACS session ID: %d, state index: %d",
		len(reqs), c.acsSessionID, c.stateOutput.GetStateIndex())
	c.workflow.setBatchProposalSent()
}

// runVMIfNeeded attempts to extract deterministic batch of requests from ACS.
// If it succeeds (i.e. all requests are available) and the extracted batch is nonempty, it runs the request
func (c *consensus) runVMIfNeeded() {
	if !c.workflow.IsConsensusBatchKnown() {
		c.log.Debugf("runVM not needed: consensus batch is not known")
		return
	}
	if c.workflow.IsVMStarted() || c.workflow.IsVMResultSigned() {
		c.log.Debugf("runVM not needed: vmStarted %v, vmResultSigned %v",
			c.workflow.IsVMStarted(), c.workflow.IsVMResultSigned())
		return
	}
	if time.Now().Before(c.delayRunVMUntil) {
		c.log.Debugf("runVM not needed: delayed till %v", c.delayRunVMUntil)
		return
	}

	reqs, missingRequestIndexes, allArrived := c.mempool.ReadyFromIDs(c.consensusBatch.TimeData, c.consensusBatch.RequestIDs...)

	c.cleanMissingRequests()

	if !allArrived {
		c.pollMissingRequests(missingRequestIndexes)
		return
	}
	if len(reqs) == 0 {
		// due to change in time, all requests became non processable ACS must be run again
		c.log.Debugf("runVM not needed: empty list of processable requests. Reset workflow")
		c.resetWorkflow()
		return
	}

	if err := c.consensusBatch.EnsureTimestampConsistent(reqs, c.stateTimestamp); err != nil {
		c.log.Errorf("Unable to ensure consistent timestamp: %v", err)
		c.resetWorkflow()
		return
	}

	c.log.Debugf("runVM needed: total number of requests = %d", len(reqs))
	// here reqs as a set is deterministic. Must be sorted to have fully deterministic list
	c.sortBatch(reqs)
	c.log.Debugf("runVM: sorted requests")

	if vmTask := c.prepareVMTask(reqs); vmTask != nil {
		chainID := iscp.ChainIDFromAliasID(vmTask.AnchorOutput.AliasID)
		c.log.Debugw("runVMIfNeeded: starting VM task",
			"chainID", (&chainID).String(),
			"milestone", vmTask.TimeAssumption.MilestoneIndex,
			"timestamp", vmTask.TimeAssumption.Time,
			"block index", vmTask.AnchorOutput.StateIndex,
			"num req", len(vmTask.Requests),
		)
		c.workflow.setVMStarted()
		c.consensusMetrics.CountVMRuns()
		go func() {
			c.vmRunner.Run(vmTask)
			if vmTask.VMError != nil {
				c.log.Errorf("runVM result: VM task failed: %v", vmTask.VMError)
				return
			}
			c.log.Debugf("runVM result: responding by state index: %d state commitment: %s",
				vmTask.VirtualStateAccess.BlockIndex(), trie.RootCommitment(vmTask.VirtualStateAccess.TrieNodeStore()))
			c.EnqueueVMResultMsg(&messages.VMResultMsg{
				Task: vmTask,
			})
			elapsed := time.Since(c.workflow.GetVMStartedTime())
			c.consensusMetrics.RecordVMRunTime(elapsed)
		}()
	} else {
		c.log.Errorf("runVM: error preparing VM task")
	}
}

func (c *consensus) pollMissingRequests(missingRequestIndexes []int) {
	// some requests are not ready, so skip VM call this time. Maybe next time will be more luck
	c.delayRunVMUntil = time.Now().Add(c.timers.VMRunRetryToWaitForReadyRequests)
	c.log.Infof( // Was silently failing when entire arrays were logged instead of counts.
		"runVM not needed: some requests didn't arrive yet. #BatchRequestIDs: %v | #BatchHashes: %v | #MissingIndexes: %v",
		len(c.consensusBatch.RequestIDs), len(c.consensusBatch.RequestHashes), len(missingRequestIndexes),
	)

	// send message to other committee nodes asking for the missing requests
	if !c.pullMissingRequestsFromCommittee {
		return
	}
	missingRequestIds := []iscp.RequestID{}
	missingRequestIDsString := ""
	for _, idx := range missingRequestIndexes {
		reqID := c.consensusBatch.RequestIDs[idx]
		reqHash := c.consensusBatch.RequestHashes[idx]
		c.missingRequestsFromBatch[reqID] = reqHash
		missingRequestIds = append(missingRequestIds, reqID)
		missingRequestIDsString += reqID.String() + ", "
	}
	c.log.Debugf("runVMIfNeeded: asking for missing requests, ids: [%v]", missingRequestIDsString)
	msg := &messages.MissingRequestIDsMsg{IDs: missingRequestIds}
	c.committeePeerGroup.SendMsgBroadcast(peering.PeerMessageReceiverChain, chain.PeerMsgTypeMissingRequestIDs, msg.Bytes())
}

// sortBatch deterministically sorts batch based on the value extracted from the consensus entropy
// It is needed for determinism and as a MEV prevention measure see [prevent-mev.md]
func (c *consensus) sortBatch(reqs []iscp.Request) {
	if len(reqs) <= 1 {
		return
	}
	rnd := util.MustUint32From4Bytes(c.consensusEntropy[:4])

	type sortStru struct {
		num uint32
		req iscp.Request
	}
	toSort := make([]sortStru, len(reqs))
	for i, req := range reqs {
		toSort[i] = sortStru{
			num: (util.MustUint32From4Bytes(req.ID().Bytes()[:4]) + rnd) & 0x0000FFFF,
			req: req,
		}
	}
	sort.Slice(toSort, func(i, j int) bool {
		switch {
		case toSort[i].num < toSort[j].num:
			return true
		case toSort[i].num > toSort[j].num:
			return false
		default: // ==
			return bytes.Compare(toSort[i].req.ID().Bytes(), toSort[j].req.ID().Bytes()) < 0
		}
	})
	for i := range reqs {
		reqs[i] = toSort[i].req
	}
}

func (c *consensus) prepareVMTask(reqs []iscp.Request) *vm.VMTask {
	stateBaseline := c.chain.GlobalStateSync().GetSolidIndexBaseline()
	if !stateBaseline.IsValid() {
		c.log.Debugf("prepareVMTask: solid state baseline is invalid. Do not even start the VM")
		return nil
	}
	task := &vm.VMTask{
		ACSSessionID:       c.acsSessionID,
		Processors:         c.chain.Processors(),
		AnchorOutput:       c.stateOutput.GetAliasOutput(),
		AnchorOutputID:     c.stateOutput.OutputID(),
		SolidStateBaseline: stateBaseline,
		Entropy:            c.consensusEntropy,
		ValidatorFeeTarget: c.consensusBatch.FeeDestination,
		Requests:           reqs,
		TimeAssumption:     c.consensusBatch.TimeData,
		VirtualStateAccess: c.currentState.Copy(),
		Log:                c.log.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
	c.log.Debugf("prepareVMTask: VM task prepared")
	return task
}

func (c *consensus) broadcastSignedResultIfNeeded() {
	if !c.workflow.IsVMResultSigned() {
		c.log.Debugf("broadcastSignedResult not needed: vm result is not signed")
		return
	}
	acksReceived := len(c.resultSigAck)
	acksNeeded := int(c.committee.Size() - 1)
	if acksReceived >= acksNeeded {
		c.log.Debugf("broadcastSignedResult not needed: acks received from %v peers, only %v needed", acksReceived, acksNeeded)
		return
	}
	if time.Now().After(c.delaySendingSignedResult) {
		signedResult := c.resultSignatures[c.committee.OwnPeerIndex()]
		msg := &messages.SignedResultMsg{
			ChainInputID: c.stateOutput.ID(),
			EssenceHash:  signedResult.EssenceHash,
			SigShare:     signedResult.SigShare,
		}
		c.committeePeerGroup.SendMsgBroadcast(peering.PeerMessageReceiverConsensus, peerMsgTypeSignedResult, util.MustBytes(msg), c.resultSigAck...)
		c.delaySendingSignedResult = time.Now().Add(c.timers.BroadcastSignedResultRetry)

		c.log.Debugf("broadcastSignedResult: broadcasted (except to %v): essence hash: %s, chain input %s",
			c.resultSigAck, msg.EssenceHash.String(), iscp.OID(msg.ChainInputID))
	} else {
		c.log.Debugf("broadcastSignedResult not needed: delayed till %v", c.delaySendingSignedResult)
	}
}

// checkQuorum when relevant check if quorum of signatures to the own calculated result is available
// If so, it aggregates signatures and finalizes the transaction.
// Then it deterministically calculates a priority sequence among contributing nodes for posting
// the transaction to L1. The deadline por posting is set proportionally to the sequence number (deterministic)
// If the node sees the transaction of the L1 before its deadline, it cancels its posting
func (c *consensus) checkQuorum() {
	if c.workflow.IsTransactionFinalized() {
		c.log.Debugf("checkQuorum not needed: transaction already finalized")
		return
	}
	if !c.workflow.IsVMResultSigned() {
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
	sigSharesToAggregate := make([]*dss.PartialSig, len(contributors))
	invalidSignatures := false
	// TODO: can dss sig shares be verified sepparatelly?
	for i, idx := range contributors {
		/*msg, err := c.resultTxEssence.SigningMessage()
		if err != nil {
			c.log.Errorf("checkQuorum: cannot retrieve message to be signed: %v", err)
			return
		}
		err = c.committee.DKShare().VerifySigShare(msg, c.resultSignatures[idx].SigShare)
		if err != nil {
			// TODO here we are ignoring wrong signatures. In general, it means it is an attack
			//  In the future when each message will be signed by the peer's identity, the invalidity
			//  of the BLS signature means the node is misbehaving and should be punished.
			c.log.Warnf("checkQuorum: INVALID SIGNATURE from peer #%d: %v", i, err)
			invalidSignatures = true
		} else {*/
		sigSharesToAggregate[i] = c.resultSignatures[idx].SigShare
		//}
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

	// if !chainOutput.GetIsGovernanceUpdated() {
	// if it is not state controller rotation, sending message to state manager
	// Otherwise state manager is not notified
	c.writeToWAL()
	c.workflow.setCurrentStateIndex(c.resultState.BlockIndex())
	chainOutputID := chainOutput.ID()
	c.chain.StateCandidateToStateManager(c.resultState, chainOutputID)
	c.log.Debugf("checkQuorum: StateCandidateMsg sent for state index %v, approving output ID %v",
		c.resultState.BlockIndex(), iscp.OID(chainOutputID))
	//}

	// calculate deterministic and pseudo-random order and postTxDeadline among contributors
	var postSeqNumber uint16
	var permutation *util.Permutation16
	txID, err := tx.ID()
	if err != nil {
		c.log.Errorf("checkQuorum failed: cannot calculate transaction ID: %v", err)
		return
	}
	if c.iAmContributor {
		permutation = util.NewPermutation16(uint16(len(c.contributors)), txID[:])
		postSeqNumber = permutation.GetArray()[c.myContributionSeqNumber]
		c.postTxDeadline = time.Now().Add(time.Duration(postSeqNumber) * c.timers.PostTxSequenceStep)

		c.log.Debugf("checkQuorum: finalized tx %s, iAmContributor: true, postSeqNum: %d, permutation: %+v",
			iscp.TxID(txID), postSeqNumber, permutation.GetArray())
	} else {
		c.log.Debugf("checkQuorum: finalized tx %s, iAmContributor: false", iscp.TxID(txID))
	}
	c.workflow.setTransactionFinalized()
	c.pullInclusionStateDeadline = time.Now()
}

func (c *consensus) writeToWAL() {
	block, err := c.resultState.ExtractBlock()
	if err == nil {
		err = c.wal.Write(block.Bytes())
		if err != nil {
			c.log.Debugf("Error writing block to wal: %v", err)
		}
	}
}

// postTransactionIfNeeded posts a finalized transaction upon deadline unless it was evidenced on L1 before the deadline.
func (c *consensus) postTransactionIfNeeded() {
	if !c.workflow.IsTransactionFinalized() {
		c.log.Debugf("postTransaction not needed: transaction is not finalized")
		return
	}
	if !c.iAmContributor {
		// only contributors post transaction
		c.log.Debugf("postTransaction not needed: i am not a contributor")
		return
	}
	if c.workflow.IsTransactionPosted() {
		c.log.Debugf("postTransaction not needed: transaction already posted")
		return
	}
	if c.workflow.IsTransactionSeen() {
		c.log.Debugf("postTransaction not needed: transaction already seen")
		return
	}
	if time.Now().Before(c.postTxDeadline) {
		c.log.Debugf("postTransaction not needed: delayed till %v", c.postTxDeadline)
		return
	}
	stateIndex := c.resultState.BlockIndex()
	go c.nodeConn.PublishTransaction(stateIndex, c.finalTx)

	c.workflow.setTransactionPosted() // TODO: Fix it, retries should be in place for robustness.
	txID, err := c.finalTx.ID()
	if err == nil {
		c.log.Infof("postTransaction: POSTED TRANSACTION for state %v: %s, number of inputs: %d, outputs: %d",
			stateIndex, iscp.TxID(txID), len(c.finalTx.Essence.Inputs), len(c.finalTx.Essence.Outputs))
	} else {
		c.log.Warnf("postTransaction: POSTED TRANSACTION for state %v: number of inputs: %d, outputs: %d, error calculating id: %v",
			stateIndex, len(c.finalTx.Essence.Inputs), len(c.finalTx.Essence.Outputs), err)
	}
}

// pullInclusionStateIfNeeded periodic pull to know the inclusions state of the transaction. Note that pulling
// starts immediately after finalization of the transaction, not after posting it
func (c *consensus) pullInclusionStateIfNeeded() {
	if !c.workflow.IsTransactionFinalized() {
		c.log.Debugf("pullInclusionState not needed: transaction is not finalized")
		return
	}
	if c.workflow.IsTransactionSeen() {
		c.log.Debugf("pullInclusionState not needed: transaction already seen")
		return
	}
	if time.Now().Before(c.pullInclusionStateDeadline) {
		c.log.Debugf("pullInclusionState not needed: delayed till %v", c.pullInclusionStateDeadline)
		return
	}
	finalTxID, err := c.finalTx.ID()
	if err != nil {
		c.log.Panicf("pullInclusionState: cannot calculate final transaction id: %v", err)
	}
	c.nodeConn.PullTxInclusionState(*finalTxID)
	c.pullInclusionStateDeadline = time.Now().Add(c.timers.PullInclusionStateRetry)
	c.log.Debugf("pullInclusionState: request for inclusion state sent")
}

// prepareBatchProposal creates a batch proposal structure out of requests
func (c *consensus) prepareBatchProposal(reqs []iscp.Request) *BatchProposal {
	consensusManaPledge := identity.ID{}
	accessManaPledge := identity.ID{}
	feeDestination := iscp.NewAgentID(c.chain.ID().AsAddress(), 0)
	// sign state output ID. It will be used to produce unpredictable entropy in consensus
	outputID := c.stateOutput.OutputID()
	sigShare, err := c.committee.DKShare().BLSSignShare(outputID[:])
	c.assert.RequireNoError(err, fmt.Sprintf("prepareBatchProposal: signing output ID %v failed", iscp.OID(c.stateOutput.ID())))

	ret := &BatchProposal{
		ValidatorIndex:          c.committee.OwnPeerIndex(),
		StateOutputID:           c.stateOutput.ID(),
		RequestIDs:              make([]iscp.RequestID, len(reqs)),
		RequestHashes:           make([][32]byte, len(reqs)),
		TimeData:                *c.timeData,
		ConsensusManaPledge:     consensusManaPledge,
		AccessManaPledge:        accessManaPledge,
		FeeDestination:          feeDestination,
		SigShareOfStateOutputID: sigShare,
	}
	for i, req := range reqs {
		ret.RequestIDs[i] = req.ID()
		ret.RequestHashes[i] = iscp.RequestHash(req)
	}

	c.log.Debugf("prepareBatchProposal: proposal prepared")
	return ret
}

// receiveACS processed new ACS received from ACS consensus
//nolint:funlen
func (c *consensus) receiveACS(values [][]byte, sessionID uint64) {
	if c.acsSessionID != sessionID {
		c.log.Debugf("receiveACS: session id missmatch: expected %v, received %v", c.acsSessionID, sessionID)
		return
	}
	if c.workflow.IsConsensusBatchKnown() {
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
		if !prop.StateOutputID.Equals(c.stateOutput.ID()) {
			c.log.Warnf("receiveACS: ACS out of context or consensus failure: expected stateOuptudId: %v, generated stateOutputID: %v ",
				iscp.OID(c.stateOutput.ID()), iscp.OID(prop.StateOutputID))
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
			c.log.Errorf("receiveACS: duplicate contributor %v in ACS", prop.ValidatorIndex)
			c.resetWorkflow()
			return
		}
		c.log.Debugf("receiveACS: contributor %v of ACS included", prop.ValidatorIndex)
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
	inBatchIDs, inBatchHashes := calcIntersection(acs, c.committee.Size())
	if len(inBatchIDs) == 0 {
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
		RequestIDs:          inBatchIDs,
		RequestHashes:       inBatchHashes,
		TimeData:            par.timeData, // It will be possibly adjusted later, when all requests are received.
		ConsensusManaPledge: par.consensusPledge,
		AccessManaPledge:    par.accessPledge,
		FeeDestination:      par.feeDestination,
	}
	c.consensusEntropy = par.entropy

	c.iAmContributor = iAmContributor
	c.myContributionSeqNumber = myContributionSeqNumber
	c.contributors = contributors

	c.workflow.setConsensusBatchKnown()

	if c.iAmContributor {
		c.log.Debugf("receiveACS: ACS received. Contributors to ACS: %+v, iAmContributor: true, seqnr: %d, reqs: %+v",
			c.contributors, c.myContributionSeqNumber, iscp.ShortRequestIDs(c.consensusBatch.RequestIDs))
	} else {
		c.log.Debugf("receiveACS: ACS received. Contributors to ACS: %+v, iAmContributor: false, reqs: %+v",
			c.contributors, iscp.ShortRequestIDs(c.consensusBatch.RequestIDs))
	}

	c.runVMIfNeeded()
}

func (c *consensus) processTxInclusionState(msg *messages.TxInclusionStateMsg) {
	if !c.workflow.IsTransactionFinalized() {
		c.log.Debugf("processTxInclusionState: transaction not finalized -> skipping.")
		return
	}
	finalTxID, err := c.finalTx.ID()
	finalTxIDStr := iscp.TxID(finalTxID)
	if err != nil {
		c.log.Panicf("processTxInclusionState: cannot calculate final transaction id: %v", err)
	}
	if msg.TxID != *finalTxID {
		c.log.Debugf("processTxInclusionState: current transaction id %v does not match the received one %v -> skipping.",
			finalTxIDStr, iscp.TxID(&msg.TxID))
		return
	}
	switch msg.State {
	case "noTransaction":
		c.log.Debugf("processTxInclusionState: transaction id %v is not known.", finalTxIDStr)
	case "included":
		c.workflow.setTransactionSeen()
		c.workflow.setCompleted()
		c.refreshConsensusInfo()
		c.log.Debugf("processTxInclusionState: transaction id %s is included; workflow finished", finalTxIDStr)
	case "conflicitng":
		c.workflow.setTransactionSeen()
		c.log.Infof("processTxInclusionState: transaction id %s is conflicting; restarting consensus.", finalTxIDStr)
		c.resetWorkflow()
	}
}

func (c *consensus) finalizeTransaction(sigSharesToAggregate []*dss.PartialSig) (*iotago.Transaction, *iscp.AliasOutputWithID, error) {
	signingBytes, err := c.resultTxEssence.SigningMessage()
	if err != nil {
		return nil, nil, xerrors.Errorf("creating signing message failed: %v", err)
	}
	signature, err := c.committee.DKShare().DSSRecoverMasterSignature(sigSharesToAggregate, signingBytes)
	if err != nil {
		return nil, nil, xerrors.Errorf("RecoverMasterSignature fail: %w", err)
	}

	// check consistency ---------------- check if chain inputs were consumed
	chainInput := c.stateOutput.ID()
	indexChainInput := -1
	for i, inp := range c.resultTxEssence.Inputs {
		if inp.Type() == iotago.InputUTXO {
			if inp.(*iotago.UTXOInput).Equals(chainInput) {
				indexChainInput = i
				break
			}
		}
	}
	c.assert.Requiref(indexChainInput >= 0, fmt.Sprintf("finalizeTransaction: cannot find tx input for state output %v. major inconsistency", iscp.OID(c.stateOutput.ID())))
	// check consistency ---------------- end

	publicKey := c.committee.DKShare().GetSharedPublic()
	var signatureArray [ed25519.SignatureSize]byte
	copy(signatureArray[:], signature)
	signatureForUnlock := &iotago.Ed25519Signature{
		PublicKey: hashing.HashData(publicKey.AsBytes()),
		Signature: signatureArray,
	}
	tx := &iotago.Transaction{
		Essence:      c.resultTxEssence,
		UnlockBlocks: transaction.MakeSignatureAndAliasUnlockBlocks(len(c.resultTxEssence.Inputs), signatureForUnlock),
	}
	chained, err := transaction.GetAliasOutput(tx, c.chain.ID().AsAddress())
	if err != nil {
		return nil, nil, err
	}
	txID, err := tx.ID()
	if err != nil {
		return nil, nil, err
	}
	c.log.Debugf("finalizeTransaction: transaction %v finalized; approving output ID: %v", iscp.TxID(txID), iscp.OID(chained.ID()))
	return tx, chained, nil
}

func (c *consensus) setNewState(msg *messages.StateTransitionMsg) bool {
	if msg.State.BlockIndex() != msg.StateOutput.GetStateIndex() {
		// NOTE: should be a panic. However this situation may occur (and occurs) in normal circumstations:
		// 1) State manager synchronizes to state index n and passes state transmission message through event to consensus asynchronously
		// 2) Consensus is overwhelmed and receives a message after delay
		// 3) Meanwhile state manager is quick enough to synchronize to state index n+1 and commits a block of state index n+1
		// 4) Only then the consensus receives a message sent in step 1. Due to imperfect implementation of virtual state copying it thinks
		//    that state is at index n+1, however chain output is (as was transmitted) and at index n.
		// The virtual state copying (earlier called "cloning") works in a following way: it copies all the mutations, stored in buffered KVS,
		// however it obtains the same kvs object to access the database. BlockIndex method of virtual state checks if there are mutations editing
		// the index value. If so, it returns the newest value in respect to mutations. Otherwise it checks the database for newest index value.
		// In the described scenario, there are no mutations neither in step 1, nor in step 3, because just before completing state synchronization
		// all the mutations are written to the DB. However, reading the same DB in step 1 results in index n and in step 4 (after the commit of block
		// index n+1) -- in index n+1. Thus effectively the virtual state received is different than the virtual state sent.
		c.log.Errorf("consensus::setNewState: state index is inconsistent: block: #%d != chain output: #%d",
			msg.State.BlockIndex(), msg.StateOutput.GetStateIndex())
		return false
	}

	c.stateOutput = msg.StateOutput
	c.currentState = msg.State
	c.stateTimestamp = msg.StateTimestamp
	oid := msg.StateOutput.OutputID()
	c.acsSessionID = util.MustUint64From8Bytes(hashing.HashData(oid[:]).Bytes()[:8])
	r := ""
	/* TODO
	if c.stateOutput.GetIsGovernanceUpdated() {
		r = " (rotate) "
	}*/
	c.log.Debugf("SET NEW STATE #%d%s, output: %s, state commitment: %s",
		msg.StateOutput.GetStateIndex(), r, iscp.OID(msg.StateOutput.ID()), trie.RootCommitment(msg.State.TrieNodeStore()))
	c.resetWorkflow()
	return true
}

func (c *consensus) resetWorkflow() {
	for i := range c.resultSignatures {
		c.resultSignatures[i] = nil
	}
	c.acsSessionID++
	c.resultState = nil
	c.resultTxEssence = nil
	c.finalTx = nil
	c.consensusBatch = nil
	c.contributors = nil
	c.resultSigAck = c.resultSigAck[:0]
	c.workflow = newWorkflowStatus(c.stateOutput != nil, c.workflow.stateIndex)
	c.log.Debugf("Workflow reset")
}

func (c *consensus) processVMResult(result *vm.VMTask) {
	if !c.workflow.IsVMStarted() ||
		c.workflow.IsVMResultSigned() ||
		c.acsSessionID != result.ACSSessionID {
		// out of context
		c.log.Debugf("processVMResult: out of context vmStarted %v, vmResultSignedAndBroadcasted %v, expected ACS session ID %v, returned ACS session ID %v",
			c.workflow.IsVMStarted(), c.workflow.IsVMResultSigned(), c.acsSessionID, result.ACSSessionID)
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
		c.assert.Requiref(result.ResultTransactionEssence != nil, "processVMResult: result.ResultTransactionEssence != nil")
		c.resultTxEssence = result.ResultTransactionEssence
		c.resultState = result.VirtualStateAccess
	}

	signingMsg, err := c.resultTxEssence.SigningMessage()
	if err != nil {
		c.log.Errorf("processVMResult: cannot obtain signing message: %v", err)
		return
	}
	signingMsgHash := hashing.HashData(signingMsg)
	c.log.Debugf("processVMResult: signing message: %s. rotate state controller: %v", signingMsgHash, rotation)

	sigShare, err := c.committee.DKShare().DSSSignShare(signingMsg)
	c.assert.RequireNoError(err, "processVMResult: ")

	c.resultSignatures[c.committee.OwnPeerIndex()] = &messages.SignedResultMsgIn{
		SignedResultMsg: messages.SignedResultMsg{
			ChainInputID: result.AnchorOutputID.UTXOInput(),
			EssenceHash:  signingMsgHash,
			SigShare:     sigShare,
		},
		SenderIndex: c.committee.OwnPeerIndex(),
	}

	c.workflow.setVMResultSigned()

	c.log.Debugf("processVMResult signed message: %s", signingMsgHash.String())
}

func (c *consensus) makeRotateStateControllerTransaction(task *vm.VMTask) *iotago.TransactionEssence {
	c.log.Debugf("makeRotateStateControllerTransaction: %s", task.RotationAddress.Bech32(c.nodeConn.L1Params().Bech32Prefix))

	// TODO access and consensus pledge
	essence, err := rotate.MakeRotateStateControllerTransaction(
		task.RotationAddress,
		task.AnchorOutput,
		task.TimeAssumption.Time,
		identity.ID{},
		identity.ID{},
	)
	c.assert.RequireNoError(err, "makeRotateStateControllerTransaction: ")
	return essence
}

func (c *consensus) receiveSignedResult(msg *messages.SignedResultMsgIn) {
	if c.resultSignatures[msg.SenderIndex] != nil {
		if c.resultSignatures[msg.SenderIndex].EssenceHash != msg.EssenceHash ||
			!bytes.Equal(c.resultSignatures[msg.SenderIndex].SigShare.Signature[:], msg.SigShare.Signature[:]) {
			c.log.Errorf("receiveSignedResult: conflicting signed result from peer %d", msg.SenderIndex)
		} else {
			c.log.Debugf("receiveSignedResult: duplicated signed result from peer %d", msg.SenderIndex)
		}
		return
	}
	if c.stateOutput == nil {
		c.log.Warnf("receiveSignedResult: chain input ID %v received from peer %d, but state output is nil",
			msg.SenderIndex, iscp.OID(msg.ChainInputID))
		return
	}
	if !msg.ChainInputID.Equals(c.stateOutput.ID()) {
		c.log.Warnf("receiveSignedResult: wrong chain input ID from peer %d: expected %v, received %v",
			msg.SenderIndex, iscp.OID(c.stateOutput.ID()), iscp.OID(msg.ChainInputID))
		return
	}
	idx := msg.SigShare.Partial.I
	if uint16(idx) >= c.committee.Size() ||
		uint16(idx) == c.committee.OwnPeerIndex() ||
		uint16(idx) != msg.SenderIndex {
		c.log.Errorf("receiveSignedResult: wrong sig share from peer %d", msg.SenderIndex)
	} else {
		c.resultSignatures[msg.SenderIndex] = msg
		c.log.Debugf("receiveSignedResult: stored sig share from sender %d, essenceHash %v", msg.SenderIndex, msg.EssenceHash)
	}
	// send acknowledgement
	msgAck := &messages.SignedResultAckMsg{
		ChainInputID: msg.ChainInputID,
		EssenceHash:  msg.EssenceHash,
	}
	c.committeePeerGroup.SendMsgByIndex(msg.SenderIndex, peering.PeerMessageReceiverConsensus, peerMsgTypeSignedResultAck, util.MustBytes(msgAck))
}

func (c *consensus) receiveSignedResultAck(msg *messages.SignedResultAckMsgIn) {
	own := c.resultSignatures[c.committee.OwnPeerIndex()]
	if own == nil {
		c.log.Debugf("receiveSignedResultAck: ack from %v ignored, because own signature is nil", msg.SenderIndex)
		return
	}
	if msg.EssenceHash != own.EssenceHash {
		c.log.Debugf("receiveSignedResultAck: ack from %v ignored, because essence hash in ack %v is different than own signature essence hash %v",
			msg.SenderIndex, msg.EssenceHash.String(), own.EssenceHash.String())
		return
	}
	if !msg.ChainInputID.Equals(own.ChainInputID) {
		c.log.Debugf("receiveSignedResultAck: ack from %v ignored, because chain input id in ack %v is different than own chain input id %v",
			msg.SenderIndex, iscp.OID(msg.ChainInputID), iscp.OID(own.ChainInputID))
		return
	}

	for _, i := range c.resultSigAck {
		if i == msg.SenderIndex {
			c.log.Debugf("receiveSignedResultAck: ack from %v ignored, because it has already been received", msg.SenderIndex)
			return
		}
	}
	c.resultSigAck = append(c.resultSigAck, msg.SenderIndex)
	c.log.Debugf("receiveSignedResultAck: ack from %v accepted; acks from nodes %v have already been received", msg.SenderIndex, c.resultSigAck)
}

// TODO mutex inside is not good

// ShouldReceiveMissingRequest returns whether the request is missing, if the incoming request matches the expects ID/Hash it is removed from the list
func (c *consensus) ShouldReceiveMissingRequest(req iscp.Request) bool {
	reqHash := hashing.HashData(req.Bytes())
	c.log.Debugf("ShouldReceiveMissingRequest: reqID %s, hash %v", req.ID(), reqHash)

	c.missingRequestsMutex.Lock()
	defer c.missingRequestsMutex.Unlock()

	expectedHash, exists := c.missingRequestsFromBatch[req.ID()]
	if !exists {
		return false
	}
	result := bytes.Equal(expectedHash[:], reqHash[:])
	if result {
		delete(c.missingRequestsFromBatch, req.ID())
	}
	return result
}

func (c *consensus) cleanMissingRequests() {
	c.missingRequestsMutex.Lock()
	defer c.missingRequestsMutex.Unlock()

	c.missingRequestsFromBatch = make(map[iscp.RequestID][32]byte) // reset list of missing requests
}
