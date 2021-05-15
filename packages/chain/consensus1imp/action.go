package consensus1imp

import (
	"bytes"
	"sort"
	"time"

	"github.com/iotaledger/wasp/packages/transaction"

	"golang.org/x/xerrors"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"

	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *consensusImpl) takeAction() {
	if !c.workflow.stateReceived || c.workflow.finished {
		return
	}
	c.proposeBatchIfNeeded()
	c.runVMIfNeeded()
	c.checkQuorum()
	c.postTransactionIfNeeded()
	c.pullInclusionStateIfNeeded()
}

func (c *consensusImpl) proposeBatchIfNeeded() {
	if c.workflow.batchProposalSent {
		return
	}
	if c.workflow.consensusBatchKnown {
		return
	}
	if time.Now().Before(c.delayBatchProposalUntil) {
		return
	}
	reqs := c.mempool.ReadyNow()
	if len(reqs) == 0 {
		return
	}
	c.log.Debugf("proposeBatchIfNeeded: ready len = %d", len(reqs))
	proposal := c.prepareBatchProposal(reqs)
	c.committee.RunACSConsensus(proposal.Bytes(), c.acsSessionID, func(sessionID uint64, acs [][]byte) {
		c.log.Debugf("received ACS")
		go c.chain.ReceiveMessage(&chain.AsynchronousCommonSubsetMsg{
			ProposedBatchesBin: acs,
			SessionID:          sessionID,
		})
	})

	c.log.Infof("proposed batch len = %d, ACS session ID: %d", len(reqs), c.acsSessionID)
	c.workflow.batchProposalSent = true
}

const waitReadyRequestsDelay = 500 * time.Millisecond

func (c *consensusImpl) runVMIfNeeded() {
	if !c.workflow.consensusBatchKnown {
		return
	}
	if c.workflow.vmStarted || c.workflow.vmResultSignedAndBroadcasted {
		return
	}
	if time.Now().Before(c.delayRunVMUntil) {
		return
	}
	reqs, allArrived := c.mempool.ReadyFromIDs(c.consensusBatch.Timestamp, c.consensusBatch.RequestIDs...)
	if !allArrived {
		// some requests are not ready, so skip VM call this time. Maybe next time will be more luck
		c.delayRunVMUntil = time.Now().Add(waitReadyRequestsDelay)
		c.log.Infof("runVMIfNeeded: some requests didn't arrive yet")
		return
	}
	if len(reqs) == 0 {
		// due to change in time, all requests became non processable ACS must be run again
		c.resetWorkflow()
		c.log.Debugf("empty list of processable requests. Reset workflow")
		return
	}
	// here reqs as as set is deterministic. Must be sorted to have fully deterministic list
	sort.Slice(reqs, func(i, j int) bool {
		switch {
		case reqs[i].Order() < reqs[j].Order():
			return true
		case reqs[i].Order() > reqs[j].Order():
			return false
		default:
			return bytes.Compare(reqs[i].ID().Bytes(), reqs[j].ID().Bytes()) < 0
		}
	})

	task := &vm.VMTask{
		ACSSessionID:       c.acsSessionID,
		Processors:         c.chain.Processors(),
		ChainInput:         c.stateOutput,
		Entropy:            hashing.HashData(c.stateOutput.ID().Bytes()),
		ValidatorFeeTarget: *c.consensusBatch.FeeDestination,
		Requests:           reqs,
		Timestamp:          c.consensusBatch.Timestamp,
		VirtualState:       c.currentState.Clone(),
		Log:                c.log,
	}
	task.OnFinish = func(_ dict.Dict, _ error, vmError error) {
		if vmError != nil {
			c.log.Errorf("VM task failed: %v", vmError)
			return
		}
		c.chain.ReceiveMessage(&chain.VMResultMsg{
			Task: task,
		})
	}

	c.workflow.vmStarted = true
	go c.vmRunner.Run(task)
}

const postSeqStepMilliseconds = 1000

func (c *consensusImpl) checkQuorum() {
	if c.workflow.transactionFinalized {
		return
	}
	if !c.workflow.vmResultSignedAndBroadcasted {
		// only can aggregate signatures if own result is calculated
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
			c.log.Warnf("wrong essence hash: expected(own): %s, got (from %d): %s", ownHash, i, sig.EssenceHash)
		}
	}
	quorumReached := len(contributors) < int(c.committee.Quorum())
	c.log.Debugf("checkQuorum: %+v, quorum = %v", contributors, quorumReached)
	if quorumReached {
		return
	}
	sigSharesToAggregate := make([][]byte, len(contributors))
	for i, idx := range contributors {
		sigSharesToAggregate[i] = c.resultSignatures[idx].SigShare
	}
	tx, approvingOutID, err := c.finalizeTransaction(sigSharesToAggregate)
	if err != nil {
		c.log.Errorf("checkQuorum: %v", err)
		return
	}

	c.finalTx = tx
	c.approvingOutputID = approvingOutID

	go c.chain.ReceiveMessage(&chain.StateCandidateMsg{
		State:             c.resultState,
		ApprovingOutputID: approvingOutID,
	})

	// calculate deterministic pseudo-random postTxDeadline among contributors
	var postSeqNumber uint16
	if c.iAmContributor {
		permutation := util.NewPermutation16(uint16(len(c.contributors)), tx.ID().Bytes())
		postSeqNumber = permutation.GetArray()[c.myContributionSeqNumber]
		c.postTxDeadline = time.Now().Add(time.Duration(postSeqNumber*postSeqStepMilliseconds) * time.Millisecond)
	}
	c.workflow.transactionFinalized = true
	c.pullInclusionStateDeadline = time.Now()
	c.log.Debugf("checkQuorum: finalized tx %s, iAmContributor: %v seqNum: %d", tx.ID().Base58(), c.iAmContributor, postSeqNumber)
}

func (c *consensusImpl) postTransactionIfNeeded() {
	if !c.workflow.transactionFinalized {
		return
	}
	if !c.iAmContributor {
		// only contributors post transaction
		return
	}
	if c.workflow.transactionPosted {
		return
	}
	if c.workflow.transactionSeen {
		return
	}
	if time.Now().Before(c.postTxDeadline) {
		return
	}
	c.nodeConn.PostTransaction(c.finalTx)

	c.workflow.transactionPosted = true
	c.log.Infof("POSTED TRANSACTION: %s", c.finalTx.ID().Base58())
}

const pullInclusionStatePeriod = 1 * time.Second

func (c *consensusImpl) pullInclusionStateIfNeeded() {
	if !c.workflow.transactionFinalized {
		return
	}
	if c.workflow.transactionSeen {
		return
	}
	if time.Now().Before(c.pullInclusionStateDeadline) {
		return
	}
	c.nodeConn.PullTransactionInclusionState(c.chain.ID().AsAddress(), c.finalTx.ID())
	c.pullInclusionStateDeadline = time.Now().Add(pullInclusionStatePeriod)
}

func (c *consensusImpl) prepareBatchProposal(reqs []coretypes.Request) *batchProposal {
	ts := time.Now()
	if !ts.After(c.stateTimestamp) {
		ts = c.stateTimestamp.Add(1 * time.Nanosecond)
	}
	consensusManaPledge := identity.ID{}
	accessManaPledge := identity.ID{}
	feeDestination := coretypes.NewAgentID(c.chain.ID().AsAddress(), 0)
	ret := &batchProposal{
		ValidatorIndex:      c.committee.OwnPeerIndex(),
		StateOutputID:       c.stateOutput.ID(),
		RequestIDs:          make([]coretypes.RequestID, len(reqs)),
		Timestamp:           ts,
		ConsensusManaPledge: consensusManaPledge,
		AccessManaPledge:    accessManaPledge,
		FeeDestination:      feeDestination,
	}
	for i := range ret.RequestIDs {
		ret.RequestIDs[i] = reqs[i].ID()
	}
	return ret
}

const delayRepeatBatchProposalFor = 500 * time.Millisecond

func (c *consensusImpl) receiveACS(values [][]byte, sessionID uint64) {
	if c.acsSessionID != sessionID {
		return
	}
	if c.workflow.consensusBatchKnown {
		// should not happen
		return
	}
	if len(values) < int(c.committee.Quorum()) {
		c.log.Errorf("receiveACS: ACS is shorter than required quorum. Ignored")
		c.resetWorkflow()
		return
	}
	acs := make([]*batchProposal, len(values))
	for i, data := range values {
		proposal, err := BatchProposalFromBytes(data)
		if err != nil {
			c.log.Errorf("receiveACS: wrong data received. Whole ACS ignored: %v", err)
			c.resetWorkflow()
			return
		}
		acs[i] = proposal
	}
	iAmContributor := false
	myContributionSeqNumber := uint16(0)
	contributors := make([]uint16, 0, c.committee.Size())
	contributorSet := make(map[uint16]struct{})
	for i, prop := range acs {
		if prop.StateOutputID != c.stateOutput.ID() {
			c.resetWorkflow()
			c.log.Warnf("receiveACS: ACS out of context or consensus failure")
			return
		}
		if prop.ValidatorIndex >= c.committee.Size() {
			c.resetWorkflow()
			c.log.Warnf("receiveACS: wrong validtor index in ACS")
			return
		}
		if prop.ValidatorIndex == c.committee.OwnPeerIndex() {
			iAmContributor = true
			myContributionSeqNumber = uint16(i)
		}
		contributors = append(contributors, prop.ValidatorIndex)
		if _, already := contributorSet[prop.ValidatorIndex]; already {
			c.resetWorkflow()
			c.log.Errorf("receiveACS: duplicate contributors in ACS")
			return
		}
		contributorSet[prop.ValidatorIndex] = struct{}{}
	}
	inBatchSet := calcIntersectionLight(acs, c.committee.Size())
	if len(inBatchSet) == 0 {
		c.log.Warnf("receiveACS: ACS intersection (light) is empty. reset workflow. State index: %d, ACS sessionID %d",
			c.stateOutput.GetStateIndex(), sessionID)
		c.resetWorkflow()
		c.delayBatchProposalUntil = time.Now().Add(delayRepeatBatchProposalFor)
		return
	}
	medianTs, accessPledge, consensusPledge, feeDestination := calcBatchParameters(acs)

	c.consensusBatch = &batchProposal{
		ValidatorIndex:      c.committee.OwnPeerIndex(),
		StateOutputID:       c.stateOutput.ID(),
		RequestIDs:          inBatchSet,
		Timestamp:           medianTs,
		ConsensusManaPledge: consensusPledge,
		AccessManaPledge:    accessPledge,
		FeeDestination:      feeDestination,
	}
	c.iAmContributor = iAmContributor
	c.myContributionSeqNumber = myContributionSeqNumber
	c.contributors = contributors

	c.workflow.consensusBatchKnown = true

	c.runVMIfNeeded()
}

func (c *consensusImpl) processInclusionState(msg *chain.InclusionStateMsg) {
	if !c.workflow.transactionFinalized {
		return
	}
	if msg.TxID != c.finalTx.ID() {
		return
	}
	switch msg.State {
	case ledgerstate.Pending:
		c.workflow.transactionSeen = true
	case ledgerstate.Confirmed:
		c.workflow.transactionSeen = true
		c.workflow.finished = true
		c.refreshConsensusInfo()
		c.log.Debugf("workflow finished. Tx confirmed: %s", msg.TxID.Base58())
	case ledgerstate.Rejected:
		c.workflow.transactionSeen = true
		c.log.Infof("transaction %s rejected. Restart consensus", msg.TxID.Base58())
		c.resetWorkflow()
	}
}

func (c *consensusImpl) finalizeTransaction(sigSharesToAggregate [][]byte) (*ledgerstate.Transaction, ledgerstate.OutputID, error) {
	signatureWithPK, err := c.committee.DKShare().RecoverFullSignature(sigSharesToAggregate, c.resultTxEssence.Bytes())
	if err != nil {
		return nil, ledgerstate.OutputID{}, xerrors.Errorf("finalizeTransaction: %w", err)
	}
	sigUnlockBlock := ledgerstate.NewSignatureUnlockBlock(ledgerstate.NewBLSSignature(*signatureWithPK))
	chainInput := ledgerstate.NewUTXOInput(c.stateOutput.ID())
	var indexChainInput = -1
	for i, inp := range c.resultTxEssence.Inputs() {
		if inp.Compare(chainInput) == 0 {
			indexChainInput = i
			break
		}
	}
	if indexChainInput < 0 {
		c.log.Panicf("RecoverFullSignature. major inconsistency")
	}

	blocks := make([]ledgerstate.UnlockBlock, len(c.resultTxEssence.Inputs()))
	for i := range c.resultTxEssence.Inputs() {
		if i == indexChainInput {
			blocks[i] = sigUnlockBlock
		} else {
			blocks[i] = ledgerstate.NewAliasUnlockBlock(uint16(indexChainInput))
		}
	}
	tx := ledgerstate.NewTransaction(c.resultTxEssence, blocks)
	approvingOutputID := transaction.GetAliasOutput(tx, c.chain.ID().AsAddress()).ID()
	return tx, approvingOutputID, nil
}

func (c *consensusImpl) setNewState(msg *chain.StateTransitionMsg) {
	c.stateOutput = msg.StateOutput
	c.currentState = msg.State
	c.stateTimestamp = msg.StateTimestamp
	c.acsSessionID = util.MustUint64From8Bytes(hashing.HashData(msg.StateOutput.ID().Bytes()).Bytes()[:8])
	c.resetWorkflow()
	c.log.Infof("SET STATE #%d, output: %s, hash: %s",
		msg.StateOutput.GetStateIndex(), coretypes.OID(msg.StateOutput.ID()), msg.State.Hash().String())
}

func (c *consensusImpl) resetWorkflow() {
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
}

func (c *consensusImpl) processVMResult(result *vm.VMTask) {
	c.log.Debugf("processVMResult")
	if !c.workflow.vmStarted ||
		c.workflow.vmResultSignedAndBroadcasted ||
		c.acsSessionID != result.ACSSessionID {
		// out of context
		return
	}
	essenceBytes := result.ResultTransactionEssence.Bytes()
	essenceHash := hashing.HashData(essenceBytes)
	c.log.Infof("VM result: essence hash: %s", essenceHash)
	sigShare, err := c.committee.DKShare().SignShare(essenceBytes)
	if err != nil {
		c.log.Panicf("processVMResult: error while signing transaction %v", err)
	}
	c.resultTxEssence = result.ResultTransactionEssence
	c.resultState = result.VirtualState
	c.resultSignatures[c.committee.OwnPeerIndex()] = &chain.SignedResultMsg{
		SenderIndex: c.committee.OwnPeerIndex(),
		EssenceHash: essenceHash,
		SigShare:    sigShare,
	}
	msg := &chain.SignedResultMsg{
		EssenceHash: essenceHash,
		SigShare:    sigShare,
	}
	c.committee.SendMsgToPeers(chain.MsgSignedResult, util.MustBytes(msg), time.Now().UnixNano())

	c.workflow.vmResultSignedAndBroadcasted = true

	c.log.Debugf("processVMResult: signed and broadcasted: essence hash: %s", msg.EssenceHash.String())
}

func (c *consensusImpl) receiveSignedResult(msg *chain.SignedResultMsg) {
	if c.resultSignatures[msg.SenderIndex] != nil {
		if c.resultSignatures[msg.SenderIndex].EssenceHash != msg.EssenceHash ||
			!bytes.Equal(c.resultSignatures[msg.SenderIndex].SigShare[:], msg.SigShare[:]) {
			c.log.Errorf("conflicting signed result from peer #%d", msg.SenderIndex)
		} else {
			c.log.Errorf("duplicated signed result from peer #%d", msg.SenderIndex)
		}
		return
	}
	idx, err := msg.SigShare.Index()
	if err != nil ||
		uint16(idx) >= c.committee.Size() ||
		uint16(idx) == c.committee.OwnPeerIndex() ||
		uint16(idx) != msg.SenderIndex {
		c.log.Errorf("wrong sig share from peer #%d", msg.SenderIndex)
		return
	}
	c.resultSignatures[msg.SenderIndex] = msg
	c.log.Debugf("stored sig share from sender %d", msg.SenderIndex)
}
