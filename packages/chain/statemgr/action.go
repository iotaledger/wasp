// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/trie"
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
	//TODO
	/*gu := ""
	if sm.stateOutput.GetIsGovernanceUpdated() {
		gu = " (rotation) "
	}
	sm.log.Debugf("notifyStateTransition: %sstate IS SYNCED to index %d and is approved by output %v",
		gu, stateOutputIndex, iscp.OID(stateOutputID))*/
	sm.log.Debugf("notifyStateTransition: state IS SYNCED to index %d and is approved by output %v",
		stateOutputIndex, iscp.OID(stateOutputID))
	sm.chain.TriggerChainTransition(&chain.ChainTransitionEventData{
		VirtualState:    sm.solidState.Copy(),
		ChainOutput:     sm.stateOutput,
		OutputTimestamp: sm.stateOutputTimestamp,
	})
}

func (sm *stateManager) isSynced() bool {
	if sm.stateOutput == nil {
		return false
	}
	l1Commitment, err := state.L1CommitmentFromAliasOutput(sm.stateOutput.GetAliasOutput())
	if err != nil {
		sm.log.Errorf("isSynced: cannot obtain state commitment from state output: %v", err)
		return false
	}
	return trie.EqualCommitments(trie.RootCommitment(sm.solidState.TrieNodeStore()), l1Commitment.StateCommitment)
}

func (sm *stateManager) pullStateIfNeeded() {
	currentTime := time.Now()
	if currentTime.After(sm.pullStateRetryTime) {
		sm.nodeConn.PullLatestOutput()
		sm.pullStateRetryTime = currentTime.Add(sm.timers.PullStateRetry)
		sm.log.Debugf("pullState: pulling state for address %s. Next pull in: %v",
			sm.chain.ID().AsAddress(), sm.pullStateRetryTime.Sub(currentTime))
	} else {
		if sm.stateOutput == nil {
			sm.log.Debugf("pullState not needed: retry in %v", sm.pullStateRetryTime.Sub(currentTime))
		} else {
			sm.log.Debugf("pullState not needed, have stateOutput.Index=%d: retry in %v",
				sm.stateOutput.GetStateIndex(), sm.pullStateRetryTime.Sub(currentTime))
		}
	}
}

func (sm *stateManager) addStateCandidateFromConsensus(nextState state.VirtualStateAccess, approvingOutputID *iotago.UTXOInput) bool {
	sm.log.Debugw("addStateCandidateFromConsensus: adding state candidate",
		"index", nextState.BlockIndex(),
		"timestamp", nextState.Timestamp(),
		"commitment", trie.RootCommitment(nextState.TrieNodeStore()),
		"output", iscp.OID(approvingOutputID),
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
	block.SetApprovingOutputID(approvingOutputID)
	sm.syncingBlocks.startSyncingIfNeeded(nextState.BlockIndex())
	sm.syncingBlocks.addBlockCandidate(block, nextState)
	sm.delayRequestBlockRetry(block.BlockIndex())

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

func (sm *stateManager) delayRequestBlockRetry(stateIndex uint32) {
	sm.syncingBlocks.setRequestBlockRetryTime(stateIndex, time.Now().Add(sm.timers.GetBlockRetry))
}

func (sm *stateManager) addBlockFromPeer(block state.Block) bool {
	sm.log.Debugf("addBlockFromPeer: adding block index %v", block.BlockIndex())
	if !sm.syncingBlocks.isSyncing(block.BlockIndex()) {
		// not asked
		sm.log.Debugf("addBlockFromPeer failed: not asked for block index %v", block.BlockIndex())
		return false
	}

	sm.syncingBlocks.addBlockCandidate(block, nil)
	if sm.syncingBlocks.hasApprovedBlockCandidate(block.BlockIndex()) { // TODO: make the timer to not spam L1
		// ask for approving output
		sm.log.Debugf("addBlockFromPeer: requesting approving output ID %v", iscp.OID(block.ApprovingOutputID()))
		sm.nodeConn.PullStateOutputByID(block.ApprovingOutputID())
	}
	return true
}

// addBlockAndCheckStateOutput function adds block to candidate list and returns true iff the block is new and is not yet approved by current stateOutput
/*func (sm *stateManager) addBlockAndCheckStateOutput(block state.Block, nextState state.VirtualStateAccess) bool {
	isBlockNew, candidate := sm.syncingBlocks.addBlockCandidate(block, nextState)
	if candidate == nil {
		return false
	}
	if isBlockNew {
		if sm.stateOutput != nil {
			sm.log.Debugf("addBlockAndCheckStateOutput: checking if block index %v (local %v, nextStateCommitment %s, approvingOutputID %v, already approved %v) is approved by current stateOutput",
				block.BlockIndex(), candidate.isLocal(), candidate.getNextStateCommitment(), iscp.OID(candidate.getApprovingOutputID()), candidate.isApproved())
			candidate.approveIfRightOutput(sm.stateOutput)
		}
		sm.log.Debugf("addBlockAndCheckStateOutput: block index %v approved %v", block.BlockIndex(), candidate.isApproved())
		return !candidate.isApproved()
	}
	return false
}*/

func (sm *stateManager) storeSyncingData() {
	if sm.stateOutput == nil {
		sm.log.Debugf("storeSyncingData not needed: stateOutput is nil")
		return
	}
	outputStateL1Commitment, err := state.L1CommitmentFromAliasOutput(sm.stateOutput.GetAliasOutput())
	if err != nil {
		sm.log.Debugf("storeSyncingData failed: error calculating stateOutput state commitment: %v", err)
		return
	}
	outputStateCommitment := outputStateL1Commitment.StateCommitment
	solidStateCommitment := trie.RootCommitment(sm.solidState.TrieNodeStore())
	sm.log.Debugf("storeSyncingData: storing values: Synced %v, SyncedBlockIndex %v, SyncedStateCommitment %s, SyncedStateTimestamp %v, StateOutputBlockIndex %v, StateOutputID %v, StateOutputCommitment %s, StateOutputTimestamp %v",
		sm.isSynced(), sm.solidState.BlockIndex(), solidStateCommitment, sm.solidState.Timestamp(), sm.stateOutput.GetStateIndex(), iscp.OID(sm.stateOutput.ID()), outputStateCommitment, sm.stateOutputTimestamp)
	sm.currentSyncData.Store(&chain.SyncInfo{
		Synced:                sm.isSynced(),
		SyncedBlockIndex:      sm.solidState.BlockIndex(),
		SyncedStateCommitment: solidStateCommitment,
		SyncedStateTimestamp:  sm.solidState.Timestamp(),
		StateOutputBlockIndex: sm.stateOutput.GetStateIndex(),
		StateOutputID:         sm.stateOutput.ID(),
		StateOutputCommitment: outputStateCommitment,
		StateOutputTimestamp:  sm.stateOutputTimestamp,
	})
}
