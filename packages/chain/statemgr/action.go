// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/wasp/packages/chain"
	"time"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

func (sm *stateManager) takeAction() {
	if !sm.ready.IsReady() {
		return
	}
	sm.checkStateApproval()
	sm.pullStateIfNeeded()
	sm.doSyncActionIfNeeded()
	sm.storeSyncingData()
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

func (sm *stateManager) checkStateApproval() {
	if sm.stateOutput == nil {
		return
	}
	// among candidate state update batches we locate the one which
	// is approved by the state output
	varStateHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		sm.log.Panic(err)
	}
	candidate, ok := sm.stateCandidates[varStateHash]
	if !ok {
		// corresponding block wasn't found among candidate state updates
		// transaction doesn't approve anything
		return
	}
	candidate.SetApprovingOutputID(sm.stateOutput.ID())

	if err := candidate.Commit(); err != nil {
		sm.log.Errorf("failed to save state at index #%d: %v", candidate.BlockIndex(), err)
	}

	sm.solidState = candidate
	sm.stateCandidates = make(map[hashing.HashValue]state.VirtualState) // clear candidate batches

	cloneState := sm.solidState.Clone()
	go sm.chain.Events().StateTransition().Trigger(&chain.StateTransitionEventData{
		VirtualState:    cloneState,
		ChainOutput:     sm.stateOutput,
		OutputTimestamp: sm.stateOutputTimestamp,
	})
	go sm.chain.Events().StateSynced().Trigger(sm.stateOutput.ID(), sm.stateOutput.GetStateIndex())
}

// adding block of state updates to the 'pending' map
func (sm *stateManager) addStateCandidate(stateCandidate state.VirtualState) {
	sm.log.Infof("added new candidate state. block index: %d, timestamp: %v",
		stateCandidate.BlockIndex(), stateCandidate.Timestamp(),
	)
	sm.stateCandidates[stateCandidate.Hash()] = stateCandidate
	sm.pullStateDeadline = time.Now()
}

func (sm *stateManager) storeSyncingData() {
	if sm.solidState == nil || sm.stateOutput == nil {
		return
	}
	outputStateHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		return
	}
	sm.currentSyncData.Store(&chain.SyncInfo{
		Synced:                sm.solidState.Hash() == outputStateHash,
		SyncedBlockIndex:      sm.solidState.BlockIndex(),
		SyncedStateHash:       sm.solidState.Hash(),
		SyncedStateTimestamp:  sm.solidState.Timestamp(),
		StateOutputBlockIndex: sm.stateOutput.GetStateIndex(),
		StateOutputID:         sm.stateOutput.ID(),
		StateOutputHash:       outputStateHash,
		StateOutputTimestamp:  sm.stateOutputTimestamp,
	})
}
