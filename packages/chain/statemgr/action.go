// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"bytes"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

func (sm *stateManager) takeAction() {
	if !sm.ready.IsReady() {
		sm.log.Debugf("takeAction skipped: state manager is not ready")
		return
	}
	sm.pullStateIfNeeded()
	sm.doSyncActionIfNeeded()
	sm.notifyStateTransitionIfNeeded()
	sm.storeSyncingData()
}

func (sm *stateManager) notifyStateTransitionIfNeeded() {
	if sm.notifiedSyncedStateHash == sm.solidState.Hash() {
		sm.log.Debugf("notifyStateTransition not needed: already notified about state %v at index #%d",
			sm.notifiedSyncedStateHash.String(), sm.solidState.BlockIndex())
		return
	}
	if !sm.isSynced() {
		sm.log.Debugf("notifyStateTransition not needed: state manager is not synced at index #%d", sm.solidState.BlockIndex())
		return
	}

	sm.notifiedSyncedStateHash = sm.solidState.Hash()
	stateOutputID := sm.stateOutput.ID()
	stateOutputIndex := sm.stateOutput.GetStateIndex()
	sm.log.Infof("notifyStateTransition: state %v IS SYNCED to index %v and is approved by output %v",
		sm.notifiedSyncedStateHash.String(), stateOutputIndex, coretypes.OID(stateOutputID))
	go sm.chain.Events().StateTransition().Trigger(&chain.StateTransitionEventData{
		VirtualState:    sm.solidState.Clone(),
		ChainOutput:     sm.stateOutput,
		OutputTimestamp: sm.stateOutputTimestamp,
	})
	go sm.chain.Events().StateSynced().Trigger(stateOutputID, stateOutputIndex)
}

func (sm *stateManager) isSynced() bool {
	if sm.stateOutput == nil {
		return false
	}
	return bytes.Equal(sm.solidState.Hash().Bytes(), sm.stateOutput.GetStateData())
}

func (sm *stateManager) pullStateIfNeeded() {
	nowis := time.Now()
	if nowis.After(sm.pullStateRetryTime) {
		chainAliasAddress := sm.chain.ID().AsAliasAddress()
		sm.nodeConn.PullState(chainAliasAddress)
		sm.pullStateRetryTime = nowis.Add(sm.timers.getPullStateRetry())
		sm.log.Debugf("pullState: pulling state for address %v. Next pull in: %v",
			chainAliasAddress.Base58(), sm.pullStateRetryTime.Sub(nowis))
	} else {
		if sm.stateOutput == nil {
			sm.log.Debugf("pullState not needed: retry in %v", sm.pullStateRetryTime.Sub(nowis))
		} else {
			sm.log.Debugf("pullState not needed at index #%d: retry in %v",
				sm.stateOutput.GetStateIndex(), sm.pullStateRetryTime.Sub(nowis))
		}
	}
}

func (sm *stateManager) addStateCandidateFromConsensus(nextState state.VirtualState, approvingOutput ledgerstate.OutputID) bool {
	sm.log.Debugw("addStateCandidateFromConsensus: adding state candidate",
		"index", nextState.BlockIndex(),
		"timestamp", nextState.Timestamp(),
		"hash", nextState.Hash(),
		"output", coretypes.OID(approvingOutput),
	)

	block, err := nextState.ExtractBlock()
	if err != nil {
		sm.log.Errorf("addStateCandidateFromConsensus: error extracting block: %v", err)
		return false
	}
	if block == nil {
		sm.log.Errorf("addStateCandidateFromConsensus: state candidate does not contain block")
		return false
	}
	block.SetApprovingOutputID(approvingOutput)
	sm.addBlockAndCheckStateOutput(block, nextState)

	if sm.stateOutput == nil || sm.stateOutput.GetStateIndex() < block.BlockIndex() {
		sm.log.Debugf("addStateCandidateFromConsensus: delaying pullStateRetry")
		sm.pullStateRetryTime = time.Now().Add(sm.timers.getPullStateNewBlockDelay())
	}

	return true
}

func (sm *stateManager) addBlockFromPeer(block state.Block) bool {
	sm.log.Debugf("addBlockFromPeer: adding block index %v", block.BlockIndex())
	if !sm.syncingBlocks.isSyncing(block.BlockIndex()) {
		// not asked
		sm.log.Debugf("addBlockFromPeer failed: not asked for block index %v", block.BlockIndex())
		return false
	}
	if sm.addBlockAndCheckStateOutput(block, nil) {
		// ask for approving output
		chainAddress := sm.chain.ID().AsAddress()
		sm.log.Debugf("addBlockFromPeer: requesting approving output ID %v for chain %v", coretypes.OID(block.ApprovingOutputID()), chainAddress.Base58())
		sm.nodeConn.PullConfirmedOutput(chainAddress, block.ApprovingOutputID())
	}
	return true
}

//addBlockAndCheckStateOutput function adds block to candidate list and returns true iff the block is new and is not yet approved by current stateOutput
func (sm *stateManager) addBlockAndCheckStateOutput(block state.Block, nextState state.VirtualState) bool {
	isBlockNew, candidate := sm.syncingBlocks.addBlockCandidate(block, nextState)
	if candidate == nil {
		return false
	}
	if isBlockNew {
		if sm.stateOutput != nil {
			sm.log.Debugf("addBlockAndCheckStateOutput: checking if block index %v (local %v, nextStateHash %v, approvingOutputID %v, already approved %v) is approved by current stateOutput",
				block.BlockIndex(), candidate.isLocal(), candidate.getNextStateHash().String(), coretypes.OID(candidate.getApprovingOutputID()), candidate.isApproved())
			candidate.approveIfRightOutput(sm.stateOutput)
		}
		sm.log.Debugf("addBlockAndCheckStateOutput: block index %v approved %v", block.BlockIndex(), candidate.isApproved())
		return !candidate.isApproved()
	}
	return false
}

func (sm *stateManager) storeSyncingData() {
	if sm.stateOutput == nil {
		sm.log.Debugf("storeSyncingData not needed: stateOutput is nil")
		return
	}
	outputStateHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		sm.log.Debugf("storeSyncingData failed: error calculating stateOutput state data hash: %v", err)
		return
	}
	sm.log.Debugf("storeSyncingData: storing values: Synced %v, SyncedBlockIndex %v, SyncedStateHash %v, SyncedStateTimestamp %v, StateOutputBlockIndex %v, StateOutputID %v, StateOutputHash %v, StateOutputTimestamp %v",
		sm.isSynced(), sm.solidState.BlockIndex(), sm.solidState.Hash().String(), sm.solidState.Timestamp(), sm.stateOutput.GetStateIndex(), coretypes.OID(sm.stateOutput.ID()), outputStateHash.String(), sm.stateOutputTimestamp)
	sm.currentSyncData.Store(&chain.SyncInfo{
		Synced:                sm.isSynced(),
		SyncedBlockIndex:      sm.solidState.BlockIndex(),
		SyncedStateHash:       sm.solidState.Hash(),
		SyncedStateTimestamp:  sm.solidState.Timestamp(),
		StateOutputBlockIndex: sm.stateOutput.GetStateIndex(),
		StateOutputID:         sm.stateOutput.ID(),
		StateOutputHash:       outputStateHash,
		StateOutputTimestamp:  sm.stateOutputTimestamp,
	})
}
