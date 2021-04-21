package statemgr

import (
	//	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

/*func (sm *stateManager) isSynced() bool {
	if sm.stateOutput == nil || sm.solidState == nil {
		return false
	}
	return bytes.Equal(sm.solidState.Hash().Bytes(), sm.stateOutput.GetStateData())
}*/

// returns block if it is known and flag if it is approved by some output
/*func (sm *stateManager) syncBlock(blockIndex uint32) (state.Block, bool) {

	if _, already := sm.syncingBlocks[blockIndex]; !already {
		sm.log.Debugf("start syncing block $%d", blockIndex)
		sm.syncingBlocks[blockIndex] = &syncingBlock{}
	}
	blk := sm.syncingBlocks[blockIndex]
	if blk.approved {
		return blk.block, true
	}
	if blk.block != nil {
		if time.Now().After(blk.pullDeadline) {
			// approval didnt come in time, probably does not exist (prunned)
			return blk.block, false
		}
		// will wait for some time more for the approval
		return nil, false
	}
	// blk.block == nil
	if time.Now().Before(blk.pullDeadline) {
		// still waiting
		return nil, false
	}
	// have to pull
	data := util.MustBytes(&chain.GetBlockMsg{
		BlockIndex: blockIndex,
	})
	// send messages until first without error
	// TODO optimize
	sm.peers.SendToAllUntilFirstError(chain.MsgGetBlock, data)
	blk.pullDeadline = time.Now().Add(periodBetweenSyncMessages)
	return nil, false
}*/

func (sm *stateManager) outputReceived(output *ledgerstate.AliasOutput) {
	if !sm.syncingBlocks.isSyncing(output.GetStateIndex()) {
		// not interested
		return
	}
	sm.syncingBlocks.approveBlockCandidates(output)
}

func (sm *stateManager) outputPushed(output *ledgerstate.AliasOutput, timestamp time.Time) {
	if sm.stateOutput != nil {
		switch {
		case sm.stateOutput.GetStateIndex() == output.GetStateIndex():
			sm.log.Debug("EventStateMsg: repeated state output")
			return
		case sm.stateOutput.GetStateIndex() > output.GetStateIndex():
			sm.log.Warn("EventStateMsg: out of order state output")
			return
		}
	}
	sm.stateOutput = output
	sm.stateOutputTimestamp = timestamp
	sm.pullStateDeadline = time.Now()
	sm.outputReceived(output)
}

func (sm *stateManager) doSyncActionIfNeeded() {
	if sm.stateOutput == nil {
		return
	}
	currentIndex := uint32(0)
	if sm.solidState != nil {
		currentIndex = sm.solidState.BlockIndex()
	}
	switch {
	case currentIndex == sm.stateOutput.GetStateIndex():
		// synced
		return
	case currentIndex > sm.stateOutput.GetStateIndex():
		sm.log.Panicf("inconsistency: solid state index is larger than state output index")
	}
	// not synced
	//TODO: is it needed?
	//if len(sm.blockCandidates) > 0 && currentIndex+1 >= sm.stateOutput.GetStateIndex() {
	//	return
	//}
	for i := currentIndex + 1; i <= sm.stateOutput.GetStateIndex(); i++ {
		if sm.syncingBlocks.getBlockCandidatesCount(i) == 0 {
			// some block is still unknown. Can't sync
			sm.requestBlockIfNeeded(i)
			return
		}
		if sm.syncingBlocks.getApprovedBlockCandidatesCount(i) > 0 {
			if sm.tryToCommitCandidates(currentIndex+1, i, true) {
				return
			}
		}
	}
	// nothing is approved but all blocks are here
	sm.tryToCommitCandidates(currentIndex+1, sm.stateOutput.GetStateIndex(), false)
}

func (sm *stateManager) requestBlockIfNeeded(stateIndex uint32) {
	if time.Now().Before(sm.syncingBlocks.getPullDeadline(stateIndex)) {
		// still waiting
		return
	}
	// have to pull
	data := util.MustBytes(&chain.GetBlockMsg{
		BlockIndex: stateIndex,
	})
	// send messages until first without error
	// TODO optimize
	sm.peers.SendToAllUntilFirstError(chain.MsgGetBlock, data)
	sm.syncingBlocks.setPullDeadline(stateIndex, time.Now().Add(periodBetweenSyncMessages))
}

func (sm *stateManager) tryToCommitCandidates(fromStateIndex uint32, toStateIndex uint32, lastStateApprovedOnly bool) bool {
	candidates, ok := sm.getCandidatesToCommit(make([]*candidateBlock, toStateIndex-fromStateIndex+1), fromStateIndex, toStateIndex, lastStateApprovedOnly)
	if ok {
		sm.commitCandidates(candidates)
	}
	return ok
}

func (sm *stateManager) getCandidatesToCommit(candidateAcc []*candidateBlock, fromStateIndex uint32, toStateIndex uint32, lastStateApprovedOnly bool) ([]*candidateBlock, bool) {
	if fromStateIndex > toStateIndex {
		//Blocks gathered. Check if the correct result is received if they are applied
		var tentativeState state.VirtualState
		if sm.solidState != nil {
			tentativeState = sm.solidState.Clone()
		} else {
			tentativeState = state.NewZeroVirtualState(sm.dbp.GetPartition(sm.chain.ID()))
		}
		for _, candidate := range candidateAcc {
			block := candidate.getBlock()
			if err := tentativeState.ApplyBlock(block); err != nil {
				sm.log.Errorf("failed to apply synced block index #%d. Error: %v", block.StateIndex(), err)
				return nil, false
			}
		}
		// state hashes must be equal
		tentativeHash := tentativeState.Hash()
		finalHash := candidateAcc[len(candidateAcc)-1].getStateHash()
		if tentativeHash != finalHash {
			sm.log.Errorf("state hashes mismatch: expected final hash: %s, tentative hash: %s", finalHash, tentativeHash)
			return nil, false
		}
		return candidateAcc, true
	}
	var stateCandidateBlocks []*candidateBlock
	if fromStateIndex == toStateIndex && lastStateApprovedOnly {
		stateCandidateBlocks = sm.syncingBlocks.getApprovedBlockCandidates(fromStateIndex)
	} else {
		stateCandidateBlocks = sm.syncingBlocks.getBlockCandidates(fromStateIndex)
	}
	//TODO: sort
	for _, stateCandidateBlock := range stateCandidateBlocks {
		resultBlocks, ok := sm.getCandidatesToCommit(append(candidateAcc, stateCandidateBlock), fromStateIndex+1, toStateIndex, lastStateApprovedOnly)
		if ok {
			return resultBlocks, true
		}
	}
	return nil, false
}

func (sm *stateManager) commitCandidates(candidates []*candidateBlock) {
	if sm.solidState == nil {
		sm.solidState = state.NewZeroVirtualState(sm.dbp.GetPartition(sm.chain.ID()))
	}
	stateIndex := uint32(0)
	for _, candidate := range candidates {
		block := candidate.getBlock()
		stateIndex := block.StateIndex()
		sm.solidState.ApplyBlock(block)
		if err := sm.solidState.CommitToDb(block); err != nil {
			sm.log.Errorf("failed to commit synced changes into DB. Restart syncing")
			sm.syncingBlocks.restartSyncing()
			return
		}
		sm.syncingBlocks.deleteSyncingBlock(stateIndex)
	}
	go sm.chain.Events().StateSynced().Trigger(candidates[len(candidates)-1].getApprovindOutputID(), stateIndex)
}

// assumes all synced already
/*func (sm *stateManager) mustCommitSynced(blocks []state.Block, finalHash hashing.HashValue, outputID ledgerstate.OutputID) {
	if len(blocks) == 0 {
		// shouldn't be here
		sm.log.Panicf("len(blocks) == 0")
	}
	var tentativeState state.VirtualState
	if sm.solidState != nil {
		tentativeState = sm.solidState.Clone()
	} else {
		tentativeState = state.NewZeroVirtualState(sm.dbp.GetPartition(sm.chain.ID()))
	}
	for _, block := range blocks {
		if err := tentativeState.ApplyBlock(block); err != nil {
			sm.log.Errorf("failed to apply synced block index #%d. Error: %v", block.StateIndex(), err)
			return
		}
	}
	// state hashes must be equal
	tentativeHash := tentativeState.Hash()
	if tentativeHash != finalHash {
		sm.log.Errorf("state hashes mismatch: expected final hash: %s, tentative hash: %s", finalHash, tentativeHash)
		return
	}
	// again applying blocks, this time seriously
	if sm.solidState == nil {
		sm.solidState = state.NewZeroVirtualState(sm.dbp.GetPartition(sm.chain.ID()))
	}
	stateIndex := uint32(0)
	for _, block := range blocks {
		stateIndex = block.StateIndex()
		sm.solidState.ApplyBlock(block)
		if err := sm.solidState.CommitToDb(block); err != nil {
			sm.log.Errorf("failed to commit synced changes into DB. Restart syncing")
			sm.syncingBlocks = make(map[uint32]*syncingBlock)
			return
		}
		delete(sm.syncingBlocks, stateIndex)
	}
	sm.blockCandidates = make(map[hashing.HashValue]*candidateBlock) // clear candidate batches
	go sm.chain.Events().StateSynced().Trigger(outputID, stateIndex)
}*/
