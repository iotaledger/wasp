package statemgr

import (
	//	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"sort"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

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

func (sm *stateManager) outputPulled(output *ledgerstate.AliasOutput) {
	sm.log.Infof("XXX outputPulled %v", coretypes.OID(output.ID()))
	if !sm.syncingBlocks.isSyncing(output.GetStateIndex()) {
		// not interested
		sm.log.Infof("XXX outputReceived: not interested %v", output.GetStateIndex())
		return
	}
	sm.syncingBlocks.approveBlockCandidates(output)
}

func (sm *stateManager) outputPushed(output *ledgerstate.AliasOutput, timestamp time.Time) {
	sm.log.Infof("XXX outputPushed %v %v %v", output.GetStateIndex(), coretypes.OID(output.ID()), timestamp)
	if sm.stateOutput != nil {
		switch {
		case sm.stateOutput.GetStateIndex() == output.GetStateIndex():
			sm.log.Infof("XXX outputPushed sm.stateOutput.GetStateIndex() == output.GetStateIndex() == %v", sm.stateOutput.GetStateIndex())
			sm.log.Debug("EventStateMsg: repeated state output")
			return
		case sm.stateOutput.GetStateIndex() > output.GetStateIndex():
			sm.log.Infof("XXX outputPushed sm.stateOutput.GetStateIndex() > output.GetStateIndex(): %v > %v", sm.stateOutput.GetStateIndex(), output.GetStateIndex())
			sm.log.Warn("EventStateMsg: out of order state output")
			return
		}
	}
	sm.stateOutput = output
	sm.stateOutputTimestamp = timestamp
	sm.syncingBlocks.approveBlockCandidates(output)
}

func (sm *stateManager) doSyncActionIfNeeded() {
	sm.log.Infof("XXX doSyncActionIfNeeded")
	if sm.stateOutput == nil {
		sm.log.Infof("XXX doSyncActionIfNeeded: output nil")
		return
	}
	switch {
	case sm.solidState.BlockIndex() == sm.stateOutput.GetStateIndex():
		// synced
		sm.log.Infof("XXX doSyncActionIfNeeded: synced")
		return
	case sm.solidState.BlockIndex() > sm.stateOutput.GetStateIndex():
		sm.log.Panicf("inconsistency: solid state index is larger than state output index")
	}
	// not synced
	//TODO: is it needed?
	//if len(sm.blockCandidates) > 0 && currentIndex+1 >= sm.stateOutput.GetStateIndex() {
	//	return
	//}
	startSyncFromIndex := sm.solidState.BlockIndex() + 1
	sm.log.Infof("XXX doSyncActionIfNeeded: from %v to %v", startSyncFromIndex, sm.stateOutput.GetStateIndex())
	for i := startSyncFromIndex; i <= sm.stateOutput.GetStateIndex(); i++ {
		//TODO: temporar if. We need to find a solution to synchronise over large gaps. Making state snapshots may help.
		if i > startSyncFromIndex+maxBlocksToCommitConst {
			go sm.chain.ReceiveMessage(chain.DismissChainMsg{
				Reason: fmt.Sprintf("StateManager.doSyncActionIfNeeded: too many blocks to catch up: %v", sm.stateOutput.GetStateIndex()-startSyncFromIndex+1)},
			)
			return
		}
		sm.log.Infof("XXX doSyncActionIfNeeded: syncing %v block candidates %v approved block candidates %v", i, sm.syncingBlocks.getBlockCandidatesCount(i), sm.syncingBlocks.getApprovedBlockCandidatesCount(i))
		if sm.syncingBlocks.getBlockCandidatesCount(i) == 0 {
			// some block is still unknown. Can't sync
			sm.requestBlockIfNeeded(i)
			return
		}
		if sm.syncingBlocks.getApprovedBlockCandidatesCount(i) > 0 {
			sm.log.Infof("XXX doSyncActionIfNeeded: tryToCommitCandidates from %v to %v", startSyncFromIndex, i)
			candidates, tentativeState, ok := sm.getCandidatesToCommit(make([]*candidateBlock, 0, i-startSyncFromIndex+1), startSyncFromIndex, i)
			if ok {
				sm.commitCandidates(candidates, tentativeState)
				return
			}
		}
	}
}

func (sm *stateManager) requestBlockIfNeeded(stateIndex uint32) {
	sm.log.Infof("XXX requestBlockIfNeeded: %v", stateIndex)
	nowis := time.Now()
	if nowis.After(sm.syncingBlocks.getRequestBlockRetryTime(stateIndex)) {
		// have to pull
		data := util.MustBytes(&chain.GetBlockMsg{
			BlockIndex: stateIndex,
		})
		// send messages until first without error
		// TODO optimize
		sm.log.Infof("XXX requestBlockIfNeeded: sending to peers")
		sm.peers.SendMsgToRandomNodes(numberOfNodesToRequestBlockFromConst, chain.MsgGetBlock, data)
		sm.syncingBlocks.startSyncingIfNeeded(stateIndex)
		sm.syncingBlocks.setRequestBlockRetryTime(stateIndex, nowis.Add(sm.timers.getGetBlockRetry()))
	} else {
		sm.log.Infof("XXX requestBlockIfNeeded: before deadline %v", sm.syncingBlocks.getRequestBlockRetryTime(stateIndex))
	}
}

func (sm *stateManager) getCandidatesToCommit(candidateAcc []*candidateBlock, fromStateIndex uint32, toStateIndex uint32) ([]*candidateBlock, state.VirtualState, bool) {
	sm.log.Infof("XXX getCandidatesToCommit: from %v to %v accumulator %v", fromStateIndex, toStateIndex, candidateAcc)
	if fromStateIndex > toStateIndex {
		//Blocks gathered. Check if the correct result is received if they are applied
		tentativeState := sm.solidState.Clone()
		for i, candidate := range candidateAcc {
			sm.log.Infof("XXX getCandidatesToCommit: candidate %v is null %v", i, candidate == nil)
			var err error
			tentativeState, err = candidate.getNextState(tentativeState)
			if err != nil {
				sm.log.Errorf("failed to apply synced block index #%d. Error: %v", candidate.getBlock().BlockIndex(), err)
				return nil, nil, false
			}
		}
		// state hashes must be equal
		tentativeHash := tentativeState.Hash()
		finalHash := candidateAcc[len(candidateAcc)-1].getNextStateHash()
		if tentativeHash != finalHash {
			sm.log.Errorf("state hashes mismatch: expected final hash: %s, tentative hash: %s", finalHash, tentativeHash)
			return nil, nil, false
		}
		return candidateAcc, tentativeState, true
	}
	var stateCandidateBlocks []*candidateBlock
	if fromStateIndex == toStateIndex {
		stateCandidateBlocks = sm.syncingBlocks.getApprovedBlockCandidates(fromStateIndex)
	} else {
		stateCandidateBlocks = sm.syncingBlocks.getBlockCandidates(fromStateIndex)
	}
	sm.log.Infof("XXX getCandidatesToCommit: stateCandidateBlocks %v", stateCandidateBlocks)
	sort.Slice(stateCandidateBlocks, func(i, j int) bool { return stateCandidateBlocks[i].getVotes() > stateCandidateBlocks[j].getVotes() })
	for _, stateCandidateBlock := range stateCandidateBlocks {
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
		sm.log.Infof("XXX will be commiting block %v", block.BlockIndex())
	}
	// TODO: very strange... You cannot commit blocks i+1, i+2, i+3,... on top of state i.
	//       You need to apply block i+1 on top of state i to receive state i+1.
	//       Only then you can commit blocks i+1, i+2, i+3,...
	//       In fact, you can commit blocks i+1, i+2, i+3,... on top of state i,
	//       but in such case a corrupted DB is received: block i+1 is commited as i+2
	//       and probably the same happens for other blocks...
	sm.solidState.ApplyBlock(blocks[0])
	//TODO: maybe commit in 10 (or some const) block batches?
	//      This would save from large commits and huge memory usage to store blocks
	err := sm.solidState.Commit(blocks...)
	if err != nil {
		sm.log.Errorf("failed to commit synced changes into DB. Restart syncing")
		sm.syncingBlocks.restartSyncing()
		return
	}
	sm.solidState = tentativeState
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
