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
	if sm.stateOutput == nil || len(sm.blockCandidates) > 0 {
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
	})
	go sm.chain.Events().StateSynced().Trigger(sm.stateOutput.ID(), sm.stateOutput.GetStateIndex())
}

// adding block of state updates to the 'pending' map
func (sm *stateManager) addBlockCandidate(block state.Block) {
	if block != nil {
		sm.log.Debugw("addBlockCandidate",
			"block index", block.BlockIndex(),
			"timestamp", block.Timestamp(),
			"size", block.Size(),
			"approving output", coretypes.OID(block.ApprovingOutputID()),
		)
	} else {
		sm.log.Debugf("addBlockCandidate: add origin candidate block")
		block = state.NewOriginBlock()
	}
	var stateToApprove state.VirtualState
	if sm.solidState == nil {
		// ignore parameter and assume original block if solidState == nil
		stateToApprove = state.CreateAndCommitOriginVirtualState(sm.dbp.GetPartition(sm.chain.ID()))
	} else {
		stateToApprove = sm.solidState.Clone()
		if err := stateToApprove.ApplyBlock(block); err != nil {
			sm.log.Error("can't apply update to the current state: %v", err)
			return
		}
	}
	// include the batch to pending batches map
	vh := stateToApprove.Hash()
	if sm.solidState == nil && vh.String() != state.OriginStateHashBase58 {
		sm.log.Panicf("major inconsistency: stateToApprove hash is %s, expected %s", vh.String(), state.OriginStateHashBase58)
	}
	sm.blockCandidates[vh] = &candidateBlock{
		block:     block,
		nextState: stateToApprove,
	}

	sm.log.Infof("added new block candidate. State index: %d, state hash: %s", block.BlockIndex(), vh.String())
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
		SyncedStateTimestamp:  time.Unix(0, sm.solidState.Timestamp()),
		StateOutputBlockIndex: sm.stateOutput.GetStateIndex(),
		StateOutputID:         sm.stateOutput.ID(),
		StateOutputHash:       outputStateHash,
		StateOutputTimestamp:  sm.stateOutputTimestamp,
	})
}
