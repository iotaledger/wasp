package statemgr

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

func (sm *stateManager) aliasOutputReceived(aliasOutput *iscp.AliasOutputWithID) bool {
	aliasOutputIndex := aliasOutput.GetStateIndex()
	aliasOutputIDStr := iscp.OID(aliasOutput.ID())
	sm.log.Debugf("aliasOutputReceived: received output index %v, id %v", aliasOutputIndex, aliasOutputIDStr)
	if sm.stateOutput == nil || sm.stateOutput.GetStateIndex() < aliasOutputIndex {
		sm.log.Debugf("aliasOutputReceived: output index %v, id %v is new state output", aliasOutputIndex, aliasOutputIDStr)
		var from uint32
		if sm.stateOutput == nil {
			from = 1
		} else {
			from = sm.stateOutput.GetStateIndex() + 1
		}
		for i := from; i <= aliasOutputIndex; i++ {
			sm.syncingBlocks.startSyncingIfNeeded(i)
		}
		sm.syncingBlocks.setApprovalInfo(aliasOutput)
		sm.stateOutput = aliasOutput
		sm.stateOutputTimestamp = time.Now()
		return true
	}
	if sm.stateOutput.GetStateIndex() == aliasOutputIndex {
		if sm.stateOutput.ID().Equals(aliasOutput.ID()) {
			sm.log.Debugf("aliasOutputReceived: output index %v, id %v is already a state output; ignored", aliasOutputIndex, aliasOutputIDStr)
			return false
		}
		// TODO implement
		/*if !output.GetIsGovernanceUpdated() {
			sm.log.Panicf("L1 inconsistency: governance transition expected in %s", iscp.OID(output.ID()))
		}*/
		// it is a state controller address rotation
		sm.log.Debugf("aliasOutputReceived:  output index %v, id %v is the same index but different ID as current state output (ID %v); it probably ir governance update output",
			aliasOutputIndex, aliasOutputIDStr, iscp.OID(sm.stateOutput.ID()))
		return false
	}
	if !sm.syncingBlocks.isSyncing(aliasOutputIndex) {
		// not interested
		sm.log.Debugf("aliasOutputReceived: state index %v is not syncing; ignoring output id %v", aliasOutputIndex, aliasOutputIDStr)
		return false
	}
	sm.log.Debugf("aliasOutputReceived: state index %v is being synced, checking if output id %v approves any blocks", aliasOutputIndex, aliasOutputIDStr)
	sm.syncingBlocks.setApprovalInfo(aliasOutput)
	return sm.syncingBlocks.hasApprovedBlockCandidate(aliasOutputIndex)
}

func (sm *stateManager) doSyncActionIfNeeded() {
	if sm.stateOutput == nil {
		sm.log.Debugf("doSyncAction not needed: stateOutput is nil")
		return
	}
	switch {
	case sm.solidState.BlockIndex() == sm.stateOutput.GetStateIndex():
		sm.log.Debugf("doSyncAction not needed: state is already synced at index #%d", sm.stateOutput.GetStateIndex())
		if sm.domain.HaveMainPeers() {
			sm.domain.SetFallbackMode(false)
		}
		return
	case sm.solidState.BlockIndex() > sm.stateOutput.GetStateIndex():
		sm.log.Debugf("BlockIndex=%v, StateIndex=%v", sm.solidState.BlockIndex(), sm.stateOutput.GetStateIndex())
		sm.log.Panicf("doSyncAction inconsistency: solid state index is larger than state output index")
	}
	// not synced
	startSyncFromIndex := sm.solidState.BlockIndex() + 1
	sm.log.Debugf("doSyncAction: trying to sync state from index %v to %v", startSyncFromIndex, sm.stateOutput.GetStateIndex())
	if !sm.domain.HaveMainPeers() || sm.syncingBlocks.blockPollFallbackNeeded() {
		sm.domain.SetFallbackMode(true)
	}
	for i := startSyncFromIndex; i <= sm.stateOutput.GetStateIndex(); i++ {
		requestBlockRetryTime := sm.syncingBlocks.getRequestBlockRetryTime(i)
		blockCandidatesCount := sm.syncingBlocks.getBlockCandidatesCount(i)
		hasApprovedBlockCandidate := sm.syncingBlocks.hasApprovedBlockCandidate(i)
		sm.log.Debugf("doSyncAction: trying to sync state for index %v; requestBlockRetryTime %v, blockCandidates count %v, has approved blockCandidate %v",
			i, requestBlockRetryTime, blockCandidatesCount, hasApprovedBlockCandidate)
		// TODO: temporary if. We need to find a solution to synchronize over large gaps. Making state snapshots may help.
		if i > startSyncFromIndex+maxBlocksToCommitConst {
			errorStr := fmt.Sprintf("StateManager.doSyncActionIfNeeded: too many blocks to catch up: %v", sm.stateOutput.GetStateIndex()-startSyncFromIndex+1)
			sm.log.Errorf(errorStr)
			sm.chain.EnqueueDismissChain(errorStr)
			return
		}
		if !sm.syncingBlocks.isObtainedFromWAL(i) && time.Now().After(requestBlockRetryTime) {
			// have to pull
			sm.log.Debugf("doSyncAction: requesting block index %v, fallback=%v from %v random peers.", i, sm.domain.GetFallbackMode(), numberOfNodesToRequestBlockFromConst)
			getBlockMsg := &messages.GetBlockMsg{BlockIndex: i}
			for _, p := range sm.domain.GetRandomOtherPeers(numberOfNodesToRequestBlockFromConst) {
				sm.domain.SendMsgByPubKey(p, peering.PeerMessageReceiverStateManager, peerMsgTypeGetBlock, util.MustBytes(getBlockMsg))
				sm.syncingBlocks.blocksPulled()
				sm.log.Debugf("doSyncAction: requesting block index %v, from %v", i, p.AsString())
			}
			sm.delayRequestBlockRetry(i)
		}
		if blockCandidatesCount == 0 {
			sm.log.Debugf("doSyncAction: no block candidates for index %v", i)
			return
		}
		if hasApprovedBlockCandidate {
			sm.log.Debugf("doSyncAction: trying to find candidates to commit from index %v to %v", startSyncFromIndex, i)
			candidates, ok := sm.getCandidatesToCommit(make([]*candidateBlock, i-startSyncFromIndex+1), startSyncFromIndex, i, sm.syncingBlocks.getApprovedBlockCandidateHash(i))
			if ok {
				sm.log.Debugf("doSyncAction: candidates to commit found, committing")
				sm.commitCandidates(candidates)
			}
			return
		}
	}
}

func (sm *stateManager) getCandidatesToCommit(candidateAcc []*candidateBlock, fromStateIndex, toStateIndex uint32, lastBlockHash state.BlockHash) ([]*candidateBlock, bool) {
	if fromStateIndex > toStateIndex {
		sm.log.Debugf("getCandidatesToCommit: all blocks found")
		return candidateAcc, true
	}
	block := sm.syncingBlocks.getBlockCandidate(toStateIndex, lastBlockHash)
	if block == nil {
		sm.log.Warnf("getCandidatesToCommit block index %v hash %s not found", toStateIndex, lastBlockHash)
		return nil, false
	}
	sm.log.Debugf("getCandidatesToCommit block index %v hash %s found", toStateIndex, lastBlockHash)
	candidateAcc[toStateIndex-fromStateIndex] = block
	return sm.getCandidatesToCommit(candidateAcc, fromStateIndex, toStateIndex-1, block.getPreviousL1Commitment().BlockHash)
}

func (sm *stateManager) commitCandidates(candidates []*candidateBlock) {
	blocks := make([]state.Block, len(candidates))
	calculatedState := sm.solidState.Copy()
	for i, candidate := range candidates {
		block := candidate.getBlock()
		blocks[i] = block
		calculatedStateCommitment := state.RootCommitment(calculatedState.TrieNodeStore())
		candidatePrevStateCommitment := block.PreviousL1Commitment().StateCommitment
		if !state.EqualCommitments(candidatePrevStateCommitment, calculatedStateCommitment) {
			sm.log.Errorf("commitCandidates: candidate index %v previous state commitment does not match calculated state commitment: %s != %s",
				block.BlockIndex(), candidatePrevStateCommitment, calculatedStateCommitment)
			sm.syncingBlocks.restartSyncing()
			return
		}
		var err error
		calculatedState, err = candidate.getNextState(calculatedState)
		calculatedState.Commit()
		if err != nil {
			sm.log.Errorf("commitCandidates: failed to apply synced block index #%d: %v",
				block.BlockIndex(), err)
			sm.syncingBlocks.restartSyncing()
			return
		}
	}

	// state commitments must be equal
	from := blocks[0].BlockIndex()
	to := blocks[len(blocks)-1].BlockIndex()
	finalStateCommitment := state.RootCommitment(calculatedState.TrieNodeStore())
	finalCandidateCommitment := sm.syncingBlocks.getNextStateCommitment(to)
	if !state.EqualCommitments(finalStateCommitment, finalCandidateCommitment) {
		sm.log.Debugf("commitCandidates: tentative state index %v obtained, however its commitment does not match last candidate expected state commitment: %s != %s",
			to, finalStateCommitment, finalCandidateCommitment)
		sm.syncingBlocks.restartSyncing()
		return
	}
	sm.log.Debugf("commitCandidates: tentative state index %v obtained, its commitment matches last candidate expected state commitment: %s",
		to, finalStateCommitment)

	for _, block := range blocks {
		sm.syncingBlocks.deleteSyncingBlock(block.BlockIndex())
		sm.log.Debugf("commitCandidates: syncing of state index %v is stopped", block.BlockIndex())
	}

	// TODO: maybe commit in 10 (or some const) block batches?
	//      This would save from large commits and huge memory usage to store blocks

	// invalidate solid state.
	// - If any VM task is running with the assumption of the previous state, it is obsolete and will self-cancel
	// - any view call will return 'state invalidated message'
	sm.chain.GlobalStateSync().InvalidateSolidIndex()
	err := calculatedState.Save(blocks...)
	for _, block := range blocks {
		sm.stateManagerMetrics.RecordBlockSize(block.BlockIndex(), float64(len(block.Bytes())))
	}

	if err != nil {
		sm.log.Errorf("commitCandidates: failed to commit synced changes into DB. Restart syncing. %w", err)
		if strings.Contains(err.Error(), "space left on device") {
			sm.log.Panicf("Terminating WASP, no space left on disc.")
		}
		sm.syncingBlocks.restartSyncing()
		return
	}
	sm.solidState = calculatedState
	sm.chain.GlobalStateSync().SetSolidIndex(sm.solidState.BlockIndex())

	sm.log.Debugf("commitCandidates: committing of block indices from %v to %v was successful", from, to)
}
