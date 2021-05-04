// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"bytes"
	"time"

	"github.com/iotaledger/wasp/packages/chain"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

func (sm *stateManager) takeAction() {
	if !sm.ready.IsReady() {
		return
	}
	sm.checkStateTransition()
	sm.notifyStateTransitionIfNeeded()
	sm.pullStateIfNeeded()
	sm.doSyncActionIfNeeded()
	sm.updateSyncingData()
}

func (sm *stateManager) pullStateIfNeeded() {
	nowis := time.Now()
	if sm.pullStateDeadline.After(nowis) {
		return
	}
	if sm.stateOutput == nil || len(sm.stateCandidates) > 0 {
		sm.log.Debugf("pull state")
		sm.nodeConn.PullState(sm.chain.ID().AsAliasAddress())
	}
	sm.pullStateDeadline = nowis.Add(pullStatePeriod)
}

func (sm *stateManager) checkStateTransition() {
	if sm.isSynced() {
		return
	}
	if sm.stateOutput == nil {
		return
	}
	// among candidate state updates we locate the one which
	// is approved by the state output
	varStateHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		sm.log.Panic(err)
	}
	candidate, ok := sm.stateCandidates[varStateHash]
	if !ok {
		// corresponding block wasn't found among stateCandidate state updates
		// transaction doesn't approve anything
		return
	}
	candidate.block.SetApprovingOutputID(sm.stateOutput.ID())

	if err := candidate.state.Commit(candidate.block); err != nil {
		sm.log.Errorf("failed to save state at index #%d: %v", candidate.state.BlockIndex(), err)
		return
	}

	sm.solidState = candidate.state
	sm.stateCandidates = make(map[hashing.HashValue]*stateCandidate) // clear
}

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
	sm.notifiedSyncedStateHash = sm.solidState.Hash()
	go sm.chain.Events().StateTransition().Trigger(&chain.StateTransitionEventData{
		VirtualState:    sm.solidState.Clone(),
		ChainOutput:     sm.stateOutput,
		OutputTimestamp: sm.stateOutputTimestamp,
	})
	go sm.chain.Events().StateSynced().Trigger(sm.stateOutput.ID(), sm.stateOutput.GetStateIndex())
}

// addStateCandidate adding state candidate of state updates to the 'pending' map. Assumes it contains the block in the log of updates
func (sm *stateManager) addStateCandidate(candidate state.VirtualState) bool {
	block, err := candidate.ExtractBlock()
	if err != nil {
		sm.log.Errorf("addStateCandidate: %v", err)
		return false
	}
	if block == nil {
		sm.log.Errorf("addStateCandidate: state candidate does not contain block")
		return false
	}
	sm.log.Infof("added new candidate state. block index: %d, timestamp: %v",
		candidate.BlockIndex(), candidate.Timestamp(),
	)
	sm.stateCandidates[candidate.Hash()] = &stateCandidate{
		state: candidate,
		block: block,
	}
	sm.pullStateDeadline = time.Now()
	return true
}

func (sm *stateManager) updateSyncingData() {
	if sm.stateOutput == nil {
		return
	}
	outputStateHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		// should not happen
		return
	}
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
