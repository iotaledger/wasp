package statemgr

import (
	"fmt"
	"sort"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

func (sm *stateManager) outputPulled(output *ledgerstate.AliasOutput) bool {
	sm.log.Debugf("outputPulled: output index %v id %v", output.GetStateIndex(), coretypes.OID(output.ID()))
	if !sm.syncingBlocks.isSyncing(output.GetStateIndex()) {
		// not interested
		sm.log.Debugf("outputPulled: not interested in output for state index %v", output.GetStateIndex())
		return false
	}
	return sm.syncingBlocks.approveBlockCandidates(output)
}

func (sm *stateManager) outputPushed(output *ledgerstate.AliasOutput, timestamp time.Time) bool {
	sm.log.Debugf("outputPushed: output index %v id %v timestampe %v", output.GetStateIndex(), coretypes.OID(output.ID()), timestamp)
	if sm.stateOutput != nil {
		switch {
		case sm.stateOutput.GetStateIndex() == output.GetStateIndex():
			sm.log.Debugf("outputPushed ignoring: repeated state output")
			return false
		case sm.stateOutput.GetStateIndex() > output.GetStateIndex():
			sm.log.Warnf("outputPushed: out of order state output; stateOutput index is already larger: %v", sm.stateOutput.GetStateIndex())
			return false
		}
	}
	sm.stateOutput = output
	sm.stateOutputTimestamp = timestamp
	sm.log.Debugf("outputPushed: stateOutput set to index %v id %v timestampe %v", output.GetStateIndex(), coretypes.OID(output.ID()), timestamp)
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
		sm.log.Debugf("doSyncAction not needed: state is already synced")
		return
	case sm.solidState.BlockIndex() > sm.stateOutput.GetStateIndex():
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
		//TODO: temporar if. We need to find a solution to synchronise over large gaps. Making state snapshots may help.
		if i > startSyncFromIndex+maxBlocksToCommitConst {
			go sm.chain.ReceiveMessage(chain.DismissChainMsg{
				Reason: fmt.Sprintf("StateManager.doSyncActionIfNeeded: too many blocks to catch up: %v", sm.stateOutput.GetStateIndex()-startSyncFromIndex+1)},
			)
			return
		}
		nowis := time.Now()
		if nowis.After(requestBlockRetryTime) {
			// have to pull
			sm.log.Debugf("doSyncAction: requesting block index %v from %v random peers", i, numberOfNodesToRequestBlockFromConst)
			data := util.MustBytes(&chain.GetBlockMsg{
				BlockIndex: i,
			})
			sm.peers.SendMsgToRandomPeersSimple(numberOfNodesToRequestBlockFromConst, chain.MsgGetBlock, data)
			sm.syncingBlocks.startSyncingIfNeeded(i)
			sm.syncingBlocks.setRequestBlockRetryTime(i, nowis.Add(sm.timers.getGetBlockRetry()))
			if blockCandidatesCount == 0 {
				return
			}
		}
		if approvedBlockCandidatesCount > 0 {
			sm.log.Debugf("doSyncAction: trying to find candidates to commit from index %v to %v", startSyncFromIndex, i)
			candidates, tentativeState, ok := sm.getCandidatesToCommit(make([]*candidateBlock, 0, i-startSyncFromIndex+1), startSyncFromIndex, i)
			if ok {
				sm.log.Debugf("doSyncAction: candidates to commit found, committing")
				sm.commitCandidates(candidates, tentativeState)
				sm.log.Debugf("doSyncAction: blocks from index %v to %v committed", startSyncFromIndex, i)
				return
			}
		}
	}
}

func (sm *stateManager) getCandidatesToCommit(candidateAcc []*candidateBlock, fromStateIndex uint32, toStateIndex uint32) ([]*candidateBlock, state.VirtualState, bool) {
	sm.log.Debugf("getCandidatesToCommit from %v to %v", fromStateIndex, toStateIndex)
	if fromStateIndex > toStateIndex {
		//Blocks gathered. Check if the correct result is received if they are applied
		tentativeState := sm.solidState.Clone()
		for _, candidate := range candidateAcc {
			var err error
			tentativeState, err = candidate.getNextState(tentativeState)
			if err != nil {
				sm.log.Errorf("getCandidatesToCommit from %v to %v: failed to apply synced block index #%d: %v",
					fromStateIndex, toStateIndex, candidate.getBlock().BlockIndex(), err)
				return nil, nil, false
			}
		}
		// state hashes must be equal
		tentativeHash := tentativeState.Hash()
		finalHash := candidateAcc[len(candidateAcc)-1].getNextStateHash()
		if tentativeHash != finalHash {
			sm.log.Debugf("getCandidatesToCommit from %v to %v: tentative state obtained, however its hash does not match last candidate expected hash: %v != %v",
				fromStateIndex, toStateIndex, tentativeHash.String(), finalHash.String())
			return nil, nil, false
		}
		sm.log.Debugf("getCandidatesToCommit from %v to %v: tentative state obtained, its hash matches last candidate expected hash: %v",
			fromStateIndex, toStateIndex, tentativeHash.String())
		return candidateAcc, tentativeState, true
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
		resultBlocks, tentativeState, ok := sm.getCandidatesToCommit(append(candidateAcc, stateCandidateBlock), fromStateIndex+1, toStateIndex)
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
	//TODO: maybe commit in 10 (or some const) block batches?
	//      This would save from large commits and huge memory usage to store blocks
	err := tentativeState.Commit(blocks...)
	if err != nil {
		sm.log.Errorf("commitCandidates: failed to commit synced changes into DB. Restart syncing")
		sm.syncingBlocks.restartSyncing()
		return
	}
	sm.solidState = tentativeState
	sm.log.Infof("commitCandidates: committing of blocks indexes from %v to %v was successful", from, to)
}
