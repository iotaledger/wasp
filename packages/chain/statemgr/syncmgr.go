package statemgr

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

func (sm *stateManager) outputPulled(output *ledgerstate.AliasOutput) bool {
	sm.log.Debugf("outputPulled: output index %v id %v", output.GetStateIndex(), iscp.OID(output.ID()))
	if !sm.syncingBlocks.isSyncing(output.GetStateIndex()) {
		// not interested
		sm.log.Debugf("outputPulled: not interested in output for state index %v", output.GetStateIndex())
		return false
	}
	return sm.syncingBlocks.approveBlockCandidates(output)
}

func (sm *stateManager) stateOutputReceived(output *ledgerstate.AliasOutput, timestamp time.Time) bool {
	sm.log.Debugf("stateOutputReceived: received output index: %v, id: %v, timestamp: %v",
		output.GetStateIndex(), iscp.OID(output.ID()), timestamp)
	if sm.solidState.BlockIndex() > output.GetStateIndex() {
		sm.log.Warnf("stateOutputReceived: out of order state output: state manager is already at state %v", sm.solidState.BlockIndex())
		return false
	}
	if sm.stateOutput != nil {
		switch {
		case sm.stateOutput.GetStateIndex() == output.GetStateIndex():
			if sm.stateOutput.ID() == output.ID() {
				// it is just a duplicate
				sm.log.Debugf("stateOutputReceived ignoring: repeated state output")
				return false
			}
			if !output.GetIsGovernanceUpdated() {
				sm.log.Panicf("L1 inconsistency: governance transition expected in %s", iscp.OID(output.ID()))
			}
			// it is a state controller address rotation

		case sm.stateOutput.GetStateIndex() > output.GetStateIndex():
			sm.log.Warnf("stateOutputReceived: out of order state output: stateOutput index is already larger: %v", sm.stateOutput.GetStateIndex())
			return false
		}
	}
	sm.stateOutput = output
	sm.stateOutputTimestamp = timestamp
	sm.log.Debugf("stateOutputReceived: stateOutput set to index %v id %v timestamp %v", output.GetStateIndex(), iscp.OID(output.ID()), timestamp)
	sm.syncingBlocks.approveBlockCandidates(output)
	return true
}

func (sm *stateManager) doSyncActionIfNeeded() {
	if sm.stateOutput == nil {
		sm.log.Debugf("doSyncAction not needed: stateOutput is nil")
		return
	}
	switch {
	case sm.solidState.BlockIndex() == sm.stateOutput.GetStateIndex():
		sm.log.Debugf("doSyncAction not needed: state is already synced at index #%d", sm.stateOutput.GetStateIndex())
		return
	case sm.solidState.BlockIndex() > sm.stateOutput.GetStateIndex():
		sm.log.Debugf("BlockIndex=%v, StateIndex=%v", sm.solidState.BlockIndex(), sm.stateOutput.GetStateIndex())
		sm.log.Panicf("doSyncAction inconsistency: solid state index is larger than state output index")
	}
	// not synced
	startSyncFromIndex := sm.solidState.BlockIndex() + 1
	sm.log.Debugf("doSyncAction: trying to sync state from index %v to %v", startSyncFromIndex, sm.stateOutput.GetStateIndex())
	for i := startSyncFromIndex; i <= sm.stateOutput.GetStateIndex(); i++ {
		requestBlockRetryTime := sm.syncingBlocks.getRequestBlockRetryTime(i)
		blockCandidatesCount := sm.syncingBlocks.getBlockCandidatesCount(i)
		approvedBlockCandidatesCount := sm.syncingBlocks.getApprovedBlockCandidatesCount(i)
		sm.log.Debugf("doSyncAction: trying to sync state for index %v; requestBlockRetryTime %v, blockCandidates count %v, approved blockCandidates count %v",
			i, requestBlockRetryTime, blockCandidatesCount, approvedBlockCandidatesCount)
		// TODO: temporary if. We need to find a solution to synchronize over large gaps. Making state snapshots may help.
		if i > startSyncFromIndex+maxBlocksToCommitConst {
			go sm.chain.ReceiveMessage(messages.DismissChainMsg{
				Reason: fmt.Sprintf("StateManager.doSyncActionIfNeeded: too many blocks to catch up: %v", sm.stateOutput.GetStateIndex()-startSyncFromIndex+1),
			},
			)
			return
		}
		nowis := time.Now()
		if nowis.After(requestBlockRetryTime) {
			// have to pull
			sm.log.Debugf("doSyncAction: requesting block index %v from %v random peers", i, numberOfNodesToRequestBlockFromConst)
			data := util.MustBytes(&messages.GetBlockMsg{
				BlockIndex: i,
			})
			sm.peers.SendMsgToRandomPeersSimple(numberOfNodesToRequestBlockFromConst, messages.MsgGetBlock, data)
			sm.syncingBlocks.startSyncingIfNeeded(i)
			sm.syncingBlocks.setRequestBlockRetryTime(i, nowis.Add(sm.timers.GetBlockRetry))
			if blockCandidatesCount == 0 {
				return
			}
		}
		if approvedBlockCandidatesCount > 0 {
			sm.log.Debugf("doSyncAction: trying to find candidates to commit from index %v to %v", startSyncFromIndex, i)
			candidates, tentativeState, ok := sm.getCandidatesToCommit(make([]*candidateBlock, 0, i-startSyncFromIndex+1), sm.solidState.Clone(), startSyncFromIndex, i)
			if ok {
				sm.log.Debugf("doSyncAction: candidates to commit found, committing")
				sm.commitCandidates(candidates, tentativeState)
				sm.log.Debugf("doSyncAction: blocks from index %v to %v committed", startSyncFromIndex, i)
				return
			}
		}
	}
}

func (sm *stateManager) getCandidatesToCommit(candidateAcc []*candidateBlock, calculatedPrevState state.VirtualState, fromStateIndex, toStateIndex uint32) ([]*candidateBlock, state.VirtualState, bool) {
	sm.log.Debugf("getCandidatesToCommit from %v to %v", fromStateIndex, toStateIndex)
	if fromStateIndex > toStateIndex {
		// state hashes must be equal
		finalStateHash := calculatedPrevState.StateCommitment()
		finalCandidateHash := candidateAcc[len(candidateAcc)-1].getNextStateHash()
		if finalStateHash != finalCandidateHash {
			sm.log.Debugf("getCandidatesToCommit from %v to %v: tentative state obtained, however its hash does not match last candidate expected hash: %v != %v",
				fromStateIndex, toStateIndex, finalStateHash.String(), finalCandidateHash.String())
			return nil, nil, false
		}
		sm.log.Debugf("getCandidatesToCommit from %v to %v: tentative state obtained, its hash matches last candidate expected hash: %v",
			fromStateIndex, toStateIndex, finalStateHash.String())
		return candidateAcc, calculatedPrevState, true
	}

	var stateCandidateBlocks []*candidateBlock
	if fromStateIndex == toStateIndex {
		stateCandidateBlocks = sm.syncingBlocks.getApprovedBlockCandidates(fromStateIndex)
	} else {
		stateCandidateBlocks = sm.syncingBlocks.getBlockCandidates(fromStateIndex)
	}
	sort.Slice(stateCandidateBlocks, func(i, j int) bool {
		return stateCandidateBlocks[i].getVotes() > stateCandidateBlocks[j].getVotes()
	})

	for i, stateCandidateBlock := range stateCandidateBlocks {
		sm.log.Debugf("getCandidatesToCommit from %v to %v: checking block %v of %v", fromStateIndex, toStateIndex, i+1, len(stateCandidateBlocks))
		candidatePrevStateHash := stateCandidateBlock.getBlock().PreviousStateHash()
		calculatedPrevStateHash := calculatedPrevState.StateCommitment()
		if candidatePrevStateHash != calculatedPrevStateHash {
			sm.log.Errorf("getCandidatesToCommit from %v to %v: candidate previous state hash does not match calculated state hash: %v <> %v",
				fromStateIndex, toStateIndex, candidatePrevStateHash.String(), calculatedPrevStateHash.String())
			return nil, nil, false
		}
		calculatedState, err := stateCandidateBlock.getNextState(calculatedPrevState)
		if err != nil {
			sm.log.Errorf("getCandidatesToCommit from %v to %v: failed to apply synced block index #%d: %v",
				fromStateIndex, toStateIndex, stateCandidateBlock.getBlock().BlockIndex(), err)
			return nil, nil, false
		}
		resultBlocks, tentativeState, ok := sm.getCandidatesToCommit(append(candidateAcc, stateCandidateBlock), calculatedState, fromStateIndex+1, toStateIndex)
		if ok {
			return resultBlocks, tentativeState, true
		}
	}
	return nil, nil, false
}

func (sm *stateManager) commitCandidates(candidates []*candidateBlock, tentativeState state.VirtualState) {
	blocks := make([]state.Block, len(candidates))
	for i, candidate := range candidates {
		block := candidate.getBlock()
		blocks[i] = block
		sm.syncingBlocks.deleteSyncingBlock(block.BlockIndex())
	}
	from := blocks[0].BlockIndex()
	to := blocks[len(blocks)-1].BlockIndex()
	sm.log.Debugf("commitCandidates: syncing of state indexes from %v to %v is stopped", from, to)

	// TODO: maybe commit in 10 (or some const) block batches?
	//      This would save from large commits and huge memory usage to store blocks

	// invalidate solid state.
	// - If any VM task is running with the assumption of the previous state, it is obsolete and will self-cancel
	// - any view call will return 'state invalidated message'
	sm.chain.GlobalStateSync().InvalidateSolidIndex()
	err := tentativeState.Commit(blocks...)
	sm.chain.GlobalStateSync().SetSolidIndex(tentativeState.BlockIndex())

	if err != nil {
		sm.log.Errorf("commitCandidates: failed to commit synced changes into DB. Restart syncing. %w", err)
		if strings.Contains(err.Error(), "space left on device") {
			sm.log.Panicf("Terminating WASP, no space left on disc.")
		}
		sm.syncingBlocks.restartSyncing()
		return
	}
	sm.solidState = tentativeState

	sm.log.Debugf("commitCandidates: committing of block indices from %v to %v was successful", from, to)
}
