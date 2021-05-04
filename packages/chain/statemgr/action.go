// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"bytes"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

func (sm *stateManager) takeAction() {
	if !sm.ready.IsReady() {
		return
	}
	//sm.checkStateApproval()
	sm.notifyStateTransitionIfNeeded()
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
			sm.pullStateRetryTime = nowis.Add(sm.timers.getPullStateRetry())
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

func (sm *stateManager) isSynced() bool {
	if sm.stateOutput == nil {
		return false
	}
	return bytes.Equal(sm.solidState.Hash().Bytes(), sm.stateOutput.GetStateData())
}

func (sm *stateManager) notifyStateTransitionIfNeeded() {
	if sm.notifiedSyncedStateHash == sm.solidState.Hash() {
		return
	}
	if !sm.isSynced() {
		return
	}

	var stateIndex uint32
	var outputID ledgerstate.OutputID
	if sm.stateOutput != nil {
		stateIndex = sm.stateOutput.GetStateIndex()
		outputID = sm.stateOutput.ID()
	}
	sm.notifiedSyncedStateHash = sm.solidState.Hash()
	go sm.chain.Events().StateTransition().Trigger(&chain.StateTransitionEventData{
		VirtualState:    sm.solidState.Clone(),
		ChainOutput:     sm.stateOutput,
		OutputTimestamp: sm.stateOutputTimestamp,
	})
	go sm.chain.Events().StateSynced().Trigger(outputID, stateIndex)
}

func (sm *stateManager) addBlockFromCommitee(nextState state.VirtualState) {
	sm.log.Infow("XXX addBlockFromCommitee",
		"block index", nextState.BlockIndex(),
		"timestamp", nextState.Timestamp(),
		"hash", nextState.Hash(),
	)
	sm.log.Debugw("addBlockCandidate",
		"block index", nextState.BlockIndex(),
		"timestamp", nextState.Timestamp(),
		"hash", nextState.Hash(),
	)

	block, err := nextState.ExtractBlock()
	if err != nil {
		sm.log.Errorf("addBlockFromCommitee: %v", err)
		return
	}
	if block == nil {
		sm.log.Errorf("addBlockFromCommitee: state candidate does not contain block")
		return
	}

	sm.addBlockAndCheckStateOutput(block, nextState)

	if sm.stateOutput == nil || sm.stateOutput.GetStateIndex() < block.BlockIndex() {
		sm.pullStateRetryTime = time.Now().Add(sm.timers.getPullStateNewBlockDelay())
	}
}

func (sm *stateManager) addBlockFromPeer(block state.Block) {
	sm.log.Infof("XXX addBlockFromNode %v", block.BlockIndex())
	if !sm.syncingBlocks.isSyncing(block.BlockIndex()) {
		// not asked
		sm.log.Infof("XXX addBlockFromNode %v: not asked", block.BlockIndex())
		return
	}
	if sm.addBlockAndCheckStateOutput(block, nil) {
		// ask for approving output
		sm.nodeConn.PullConfirmedOutput(sm.chain.ID().AsAddress(), block.ApprovingOutputID())
	}
}

func (sm *stateManager) addBlockAndCheckStateOutput(block state.Block, nextState state.VirtualState) bool {
	isBlockNew, candidate, err := sm.syncingBlocks.addBlockCandidate(block, nextState)
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
	sm.log.Infof("XXX storeSyncingData: synced %v block index %v state hash %v state timestamp %v output index %v output id %v output hash %v output timestamp %v", sm.solidState.Hash() == outputStateHash, sm.solidState.BlockIndex(), sm.solidState.Hash(), sm.solidState.Timestamp(), sm.stateOutput.GetStateIndex(), sm.stateOutput.ID(), outputStateHash, sm.stateOutputTimestamp)
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
