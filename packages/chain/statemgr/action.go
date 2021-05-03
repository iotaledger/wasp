// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/wasp/packages/chain"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

func (sm *stateManager) takeAction() {
	if !sm.ready.IsReady() {
		return
	}
	//sm.checkStateApproval()
	sm.pullStateIfNeeded()
	sm.doSyncActionIfNeeded()
	sm.storeSyncingData()
}

func (sm *stateManager) pullStateIfNeeded() {
	sm.log.Infof("XXX pullStateIfNeeded")
	nowis := time.Now()
	if nowis.After(sm.pullStateRetryTime) {
		if sm.stateOutput == nil || sm.syncingBlocks.hasBlockCandidatesNotOlderThan(sm.stateOutput.GetStateIndex()+1) {
			sm.log.Infof("XXX pullStateIfNeeded: pull it")
			sm.log.Debugf("pull state")
			sm.nodeConn.PullState(sm.chain.ID().AsAliasAddress())
			sm.pullStateRetryTime = nowis.Add(pullStateRetryConst)
		}
	} else {
		sm.log.Infof("XXX pullStateIfNeeded: before retry time")
	}
}

/*func (sm *stateManager) checkStateApproval() {
	if sm.stateOutput == nil {
		return
	}
	// among candidate state update batches we locate the one which
	// is approved by the state output
	varStateHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		sm.log.Panic(err)
	}
	candidate, ok := sm.blockCandidates[varStateHash]
	if !ok {
		// corresponding block wasn't found among candidate state updates
		// transaction doesn't approve anything
		return
	}

	// found a candidate block which is approved by the stateOutput
	// set the transaction id from output
	candidate.block.WithApprovingOutputID(sm.stateOutput.ID())
	// save state to db
	if err := candidate.nextState.CommitToDb(candidate.block); err != nil {
		sm.log.Errorf("failed to save state at index #%d: %v", candidate.nextState.BlockIndex(), err)
		return
	}
	sm.solidState = candidate.nextState
	sm.blockCandidates = make(map[hashing.HashValue]*candidateBlock) // clear candidate batches

	cloneState := sm.solidState.Clone()
	go sm.chain.Events().StateTransition().Trigger(&chain.StateTransitionEventData{
		VirtualState:     cloneState,
		BlockEssenceHash: candidate.block.EssenceHash(),
		ChainOutput:      sm.stateOutput,
		Timestamp:        sm.stateOutputTimestamp,
		RequestIDs:       candidate.block.RequestIDs(),
	})
	go sm.chain.Events().StateSynced().Trigger(sm.stateOutput.ID(), sm.stateOutput.GetStateIndex())
}*/

func (sm *stateManager) addBlockFromCommitee(block state.Block) {
	sm.addBlockFromSelf(block)
	if sm.stateOutput == nil || sm.stateOutput.GetStateIndex() < block.StateIndex() {
		sm.pullStateRetryTime = time.Now().Add(pullStateNewBlockDelayConst)
	}
}

// adding block of state updates to the 'pending' map
func (sm *stateManager) addBlockFromSelf(block state.Block) {
	sm.log.Infow("XXX addBlockFromCommitee",
		"block index", block.StateIndex(),
		"timestamp", block.Timestamp(),
		"size", block.Size(),
		"approving output", coretypes.OID(block.ApprovingOutputID()),
	)
	sm.log.Debugw("addBlockCandidate",
		"block index", block.StateIndex(),
		"timestamp", block.Timestamp(),
		"size", block.Size(),
		"approving output", coretypes.OID(block.ApprovingOutputID()),
	)

	var nextState state.VirtualState
	if sm.solidState == nil {
		nextState = state.NewVirtualState(sm.dbp.GetPartition(sm.chain.ID()), nil)
	} else {
		nextState = sm.solidState.Clone()
	}
	if err := nextState.ApplyBlock(block); err != nil {
		sm.log.Error("can't apply update to the current state: %v", err)
		return
	}
	// include the batch to pending batches map
	nextStateHash := nextState.Hash()
	if sm.solidState == nil && nextStateHash.String() != state.OriginStateHashBase58 {
		sm.log.Panicf("major inconsistency: stateToApprove hash is %s, expected %s", nextStateHash.String(), state.OriginStateHashBase58)
	}

	sm.addBlockAndCheckStateOutput(block, &nextStateHash)
}

func (sm *stateManager) addBlockFromPeer(block state.Block) {
	sm.log.Infof("XXX addBlockFromNode %v", block.StateIndex())
	if !sm.syncingBlocks.isSyncing(block.StateIndex()) {
		// not asked
		sm.log.Infof("XXX addBlockFromNode %v: not asked", block.StateIndex())
		return
	}
	if sm.addBlockAndCheckStateOutput(block, nil) {
		// ask for approving output
		sm.nodeConn.PullConfirmedOutput(sm.chain.ID().AsAddress(), block.ApprovingOutputID())
	}
}

func (sm *stateManager) addBlockAndCheckStateOutput(block state.Block, stateHash *hashing.HashValue) bool {
	isBlockNew, candidate, err := sm.syncingBlocks.addBlockCandidate(block, stateHash)
	if err != nil {
		sm.log.Error(err)
		return false
	}
	if isBlockNew {
		if sm.stateOutput != nil {
			candidate.approveIfRightOutput(sm.stateOutput)
		}
		return !candidate.isApproved()
	}
	return false
}

func (sm *stateManager) storeSyncingData() {
	sm.log.Infof("XXX storeSyncingData")
	if sm.solidState == nil || sm.stateOutput == nil {
		return
	}
	outputStateHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		return
	}
	sm.log.Infof("XXX storeSyncingData: synced %v block index %v state hash %v state timestamp %v output index %v output id %v output hash %v output timestamp %v", sm.solidState.Hash() == outputStateHash, sm.solidState.BlockIndex(), sm.solidState.Hash(), time.Unix(0, sm.solidState.Timestamp()), sm.stateOutput.GetStateIndex(), sm.stateOutput.ID(), outputStateHash, sm.stateOutputTimestamp)
	sm.currentSyncData.Store(&chain.SyncInfo{
		Synced:                sm.solidState.Hash() == outputStateHash,
		SyncedBlockIndex:      sm.solidState.BlockIndex(),
		SyncedStateHash:       sm.solidState.Hash(),
		SyncedStateTimestamp:  time.Unix(0, sm.solidState.Timestamp()),
		StateOutputBlockIndex: sm.stateOutput.GetStateIndex(),
		StateOutputID:         sm.stateOutput.ID(),
		StateOutputHash:       outputStateHash,
		StateOutputTimestamp:  sm.stateOutputTimestamp,
	})
}
