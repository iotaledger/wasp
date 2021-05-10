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
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *consensusImpl) takeAction() {
	c.startConsensusIfNeeded()
	c.runVMIfNeeded()
	c.checkQuorum()
	c.postTransactionIfNeeded()
}

func (c *consensusImpl) startConsensusIfNeeded() {
	if c.stage != stageStateReceived {
		return
	}
	reqs := c.mempool.GetReadyList()
	if len(reqs) == 0 {
		return
	}
	proposal := c.prepareBatchProposal(reqs)
	c.sendProposalForACS(proposal)
}

func (c *consensusImpl) runVMIfNeeded() {
	if c.stage != stageConsensusCompleted {
		return
	}
	reqs := c.mempool.GetRequestsByIDs(c.consensusBatch.Timestamp, c.consensusBatch.RequestIDs...)
	// check if all ready
	for _, req := range reqs {
		if req == nil {
			// some requests are not ready, so skip VM call this time. Maybe next time will be luckier
			c.log.Debugf("runVMIfNeeded: not all requests ready for processing")
			return
		}
	}
	// sort for determinism and for the reason to arrange off-ledger requests
	// equal Order() request are ordered by ID
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

	c.stage = stageVM
	c.stageStarted = time.Now()

	runvm.MustRunVMTaskAsync(task)
}

const postSeqStepSeconds = 1

func (c *consensusImpl) checkQuorum() {
	if c.resultSignatures[c.committee.OwnPeerIndex()] == nil {
		// only can aggregate signatures if own result is calculated
		return
	}
	ownHash := c.resultSignatures[c.committee.OwnPeerIndex()].EssenceHash
	numSigs := uint16(0)
	for _, sig := range c.resultSignatures {
		if sig == nil {
			continue
		}
		if sig.EssenceHash == ownHash {
			numSigs++
		}
	}
	if numSigs < c.committee.Quorum() {
		return
	}
	sigSharesToAggregate := make([][]byte, 0, numSigs)
	for _, sig := range c.resultSignatures {
		if sig == nil {
			continue
		}
		if sig.EssenceHash == ownHash {
			sigSharesToAggregate = append(sigSharesToAggregate, sig.SigShare)
		}
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

	c.stage = stageTransactionFinalized
	c.stageStarted = time.Now()

	// TODO permute only those which contributed to ACS
	permutation := util.NewPermutation16(c.committee.Size(), tx.ID().Bytes())
	postSeqNumber := permutation.GetArray()[c.committee.OwnPeerIndex()]
	c.postTxDeadline = c.stageStarted.Add(time.Duration(postSeqNumber*postSeqStepSeconds) * time.Second)
}

func (c *consensusImpl) postTransactionIfNeeded() {
	if c.stage != stageTransactionFinalized {
		return
	}
	if time.Now().Before(c.postTxDeadline) {
		return
	}
	c.nodeConn.PostTransaction(c.finalTx)

	c.committee.SendMsgToPeers(chain.MsgNotifyFinalResultPosted, util.MustBytes(&chain.NotifyFinalResultPostedMsg{
		StateOutputID: c.approvingOutputID,
		TxId:          c.finalTx.ID(),
	}), time.Now().UnixNano())

	c.stage = stageWaitNextState
	c.stageStarted = time.Now()
	c.log.Infof("POSTED TRANSACTION: %s", c.finalTx.ID().Base58())

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
	for i := range c.resultSignatures {
		c.resultSignatures[i] = nil
	}
	c.resultState = nil
	c.resultTxEssence = nil
	c.finalTx = nil

	c.stage = stageStateReceived
	c.stageStarted = time.Now()
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

func (c *consensusImpl) sendProposalForACS(proposal *batchProposal) {
	c.mockedACS.ProposeValue(proposal.Bytes(), c.stateOutput.ID().Bytes())
	c.stage = stageConsensus
	c.stageStarted = time.Now()
}

func (c *consensusImpl) receiveACS(values [][]byte) {
	if len(values) < int(c.committee.Quorum()) {
		c.log.Errorf("receiveACS: ACS is shorter than equired quorum. Ignored")
		return
	}
	acs := make([]*batchProposal, len(values))
	for i, data := range values {
		proposal, err := BatchProposalFromBytes(data)
		if err != nil {
			c.log.Errorf("receiveACS: wrong data received. Whole ACS ignored: %v", err)
			return
		}
		acs[i] = proposal
	}
	for _, prop := range acs {
		if prop.StateOutputID != c.stateOutput.ID() {
			//
			c.log.Warnf("receiveACS: ACS out of context or consensus failure")
			return
		}
		if prop.ValidatorIndex >= c.committee.Size() {
			c.log.Warnf("receiveACS: wrong validtor index in ACS")
			return
		}
	}
	inBatchSet := calcIntersection(acs, c.committee.Size(), c.committee.Quorum())
	if len(inBatchSet) == 0 {
		c.log.Warnf("receiveACS: intersecection is empty. Consensus failure")
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
	c.stage = stageConsensusCompleted
	c.stageStarted = time.Now()

	c.runVMIfNeeded()
}

func (c *consensusImpl) processVMResult(result *vm.VMTask) {
	c.log.Debugw("processVMResult")

	essenceBytes := result.ResultTransactionEssence.Bytes()
	essenceHash := hashing.HashData(essenceBytes)
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

	c.stage = stageWaitForSignatures
	c.stageStarted = time.Now()

	c.log.Debugf("processVMResult: signed and broadcasted: essence hash: %s", msg.EssenceHash.String())
}

func (c *consensusImpl) processSignedResult(msg *chain.SignedResultMsg) {
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
}
