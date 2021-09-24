// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"bytes"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
)

func (sm *stateManager) takeAction() {
	if !sm.ready.IsReady() {
		sm.log.Debugf("takeAction skipped: state manager is not ready")
		return
	}
	sm.pullStateIfNeeded()
	sm.doSyncActionIfNeeded()
	sm.notifyChainTransitionIfNeeded()
	sm.storeSyncingData()
}

func (sm *stateManager) notifyChainTransitionIfNeeded() {
	if sm.stateOutput == nil {
		return
	}
	if sm.notifiedAnchorOutputID == sm.stateOutput.ID() {
		sm.log.Debugf("notifyStateTransition not needed: already notified about state %v at index #%d",
			iscp.OID(sm.notifiedAnchorOutputID), sm.solidState.BlockIndex())
		return
	}
	if !sm.isSynced() {
		sm.log.Debugf("notifyStateTransition not needed: state manager is not synced at index #%d", sm.solidState.BlockIndex())
		return
	}

	sm.notifiedAnchorOutputID = sm.stateOutput.ID()
	stateOutputID := sm.stateOutput.ID()
	stateOutputIndex := sm.stateOutput.GetStateIndex()
	gu := ""
	if sm.stateOutput.GetIsGovernanceUpdated() {
		gu = " (rotation) "
	}
	sm.log.Debugf("notifyStateTransition: %sstate IS SYNCED to index %d and is approved by output %v",
		gu, stateOutputIndex, iscp.OID(stateOutputID))
	sm.chain.Events().ChainTransition().Trigger(&chain.ChainTransitionEventData{
		VirtualState:    sm.solidState.Clone(),
		ChainOutput:     sm.stateOutput,
		OutputTimestamp: sm.stateOutputTimestamp,
	})
}

func (sm *stateManager) isSynced() bool {
	if sm.stateOutput == nil {
		return false
	}
	return bytes.Equal(sm.solidState.StateCommitment().Bytes(), sm.stateOutput.GetStateData())
}

func (sm *stateManager) pullStateIfNeeded() {
	nowis := time.Now()
	if nowis.After(sm.pullStateRetryTime) {
		chainAliasAddress := sm.chain.ID().AsAliasAddress()
		sm.nodeConn.PullState(chainAliasAddress)
		sm.pullStateRetryTime = nowis.Add(sm.timers.PullStateRetry)
		sm.log.Debugf("pullState: pulling state for address %v. Next pull in: %v",
			chainAliasAddress.Base58(), sm.pullStateRetryTime.Sub(nowis))
	} else {
		if sm.stateOutput == nil {
			sm.log.Debugf("pullState not needed: retry in %v", sm.pullStateRetryTime.Sub(nowis))
		} else {
			sm.log.Debugf("pullState not needed, have stateOutput.Index=%d: retry in %v",
				sm.stateOutput.GetStateIndex(), sm.pullStateRetryTime.Sub(nowis))
		}
	}
}

func (sm *stateManager) addStateCandidateFromConsensus(nextState state.VirtualState, approvingOutput ledgerstate.OutputID) bool {
	sm.log.Debugw("addStateCandidateFromConsensus: adding state candidate",
		"index", nextState.BlockIndex(),
		"timestamp", nextState.Timestamp(),
		"hash", nextState.StateCommitment(),
		"output", iscp.OID(approvingOutput),
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
	if sm.solidState != nil && sm.solidState.BlockIndex() >= block.BlockIndex() {
		// already processed
		sm.log.Warnf("addStateCandidateFromConsensus: block index %v is not needed as solid state is already at index %v", block.BlockIndex(), sm.solidState.BlockIndex())
		return false
	}
	block.SetApprovingOutputID(approvingOutput)
	sm.addBlockAndCheckStateOutput(block, nextState)

	if sm.stateOutput == nil || sm.stateOutput.GetStateIndex() < block.BlockIndex() {
		if sm.stateOutput == nil {
			sm.log.Debugf("addStateCandidateFromConsensus: delaying pullStateRetry for %v: state output is nil", sm.timers.PullStateAfterStateCandidateDelay)
		} else {
			sm.log.Debugf("addStateCandidateFromConsensus: delaying pullStateRetry for %v: state output index %v is less than block index %v",
				sm.timers.PullStateAfterStateCandidateDelay, sm.stateOutput.GetStateIndex(), block.BlockIndex())
		}
		sm.pullStateRetryTime = time.Now().Add(sm.timers.PullStateAfterStateCandidateDelay)
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
		sm.log.Debugf("addBlockFromPeer: requesting approving output ID %v for chain %v", iscp.OID(block.ApprovingOutputID()), chainAddress.Base58())
		sm.nodeConn.PullConfirmedOutput(chainAddress, block.ApprovingOutputID())
	}
	return true
}

// addBlockAndCheckStateOutput function adds block to candidate list and returns true iff the block is new and is not yet approved by current stateOutput
func (sm *stateManager) addBlockAndCheckStateOutput(block state.Block, nextState state.VirtualState) bool {
	isBlockNew, candidate := sm.syncingBlocks.addBlockCandidate(block, nextState)
	if candidate == nil {
		return false
	}
	if isBlockNew {
		if sm.stateOutput != nil {
			sm.log.Debugf("addBlockAndCheckStateOutput: checking if block index %v (local %v, nextStateHash %v, approvingOutputID %v, already approved %v) is approved by current stateOutput",
				block.BlockIndex(), candidate.isLocal(), candidate.getNextStateHash().String(), iscp.OID(candidate.getApprovingOutputID()), candidate.isApproved())
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
		sm.isSynced(), sm.solidState.BlockIndex(), sm.solidState.StateCommitment().String(), sm.solidState.Timestamp(), sm.stateOutput.GetStateIndex(), iscp.OID(sm.stateOutput.ID()), outputStateHash.String(), sm.stateOutputTimestamp)
	sm.currentSyncData.Store(&chain.SyncInfo{
		Synced:                sm.isSynced(),
		SyncedBlockIndex:      sm.solidState.BlockIndex(),
		SyncedStateHash:       sm.solidState.StateCommitment(),
		SyncedStateTimestamp:  sm.solidState.Timestamp(),
		StateOutputBlockIndex: sm.stateOutput.GetStateIndex(),
		StateOutputID:         sm.stateOutput.ID(),
		StateOutputHash:       outputStateHash,
		StateOutputTimestamp:  sm.stateOutputTimestamp,
	})
}
