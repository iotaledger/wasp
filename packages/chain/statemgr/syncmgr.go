package statemgr

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

// return if synced already
func (sm *stateManager) syncBlock(blockIndex uint32) state.Block {
	if _, already := sm.syncedBlocks[blockIndex]; !already {
		sm.syncedBlocks[blockIndex] = &syncingBlock{}
	}
	blk := sm.syncedBlocks[blockIndex]
	if blk.block != nil {
		return blk.block
	}
	if !blk.pullDeadline.After(time.Now()) {
		return nil
	}
	// it is time to ask for the next state update to next peer in the permutation
	data := util.MustBytes(&chain.GetBlockMsg{
		BlockIndex: blockIndex,
	})
	// send messages until first without error
	// TODO optimize
	sm.peers.SendToAllUntilFirstError(chain.MsgGetBatch, data)
	blk.pullDeadline = time.Now().Add(periodBetweenSyncMessages)
	return nil
}

func (sm *stateManager) blockHeaderArrived(msg *chain.BlockHeaderMsg) {
	syncBlk, ok := sm.syncedBlocks[msg.BlockIndex]
	if !ok {
		// not asked
		return
	}
	if syncBlk.block != nil || syncBlk.stateUpdates != nil {
		// already
		return
	}
	syncBlk.stateUpdates = make([]state.StateUpdate, msg.Size)
	syncBlk.stateTxID = msg.AnchorTransactionID
	syncBlk.pullDeadline = time.Now().Add(periodBetweenSyncMessages)
}

func (sm *stateManager) stateUpdateArrived(msg *chain.StateUpdateMsg) {
	syncBlk, ok := sm.syncedBlocks[msg.BlockIndex]
	if !ok {
		// not asked
		return
	}
	if syncBlk.block != nil {
		// already synced this block
		return
	}
	if syncBlk.stateUpdates == nil {
		// header must come first
		return
	}
	if int(msg.IndexInTheBlock) >= len(syncBlk.stateUpdates) {
		// wrong index
		return
	}
	if syncBlk.stateUpdates[msg.IndexInTheBlock] == nil {
		syncBlk.stateUpdates[msg.IndexInTheBlock] = msg.StateUpdate
		syncBlk.msgCounter++
	}
	if int(syncBlk.msgCounter) == len(syncBlk.stateUpdates) {
		block, err := state.NewBlock(syncBlk.stateUpdates...)
		if err != nil {
			sm.log.Errorf("failed to create block: %v", err)
			return
		}
		block.WithBlockIndex(msg.BlockIndex).WithStateTransaction(syncBlk.stateTxID)
		syncBlk.block = block
		syncBlk.stateUpdates = nil
	}
}

func (sm *stateManager) doSyncIfNeeded() {
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
	allSynced := true
	for i := currentIndex + 1; i < sm.stateOutput.GetStateIndex(); i++ {
		if sm.syncBlock(i) == nil {
			allSynced = false
		}
	}
	if allSynced {
		sm.mustCommitSynced(currentIndex + 1)
	}
}

// assumes all synced already
func (sm *stateManager) mustCommitSynced(fromIndex uint32) {
	// all synced, we need to push all blocks into the state
	defer func() {
		if len(sm.syncedBlocks) != 0 {
			sm.log.Panicf("inconsistency: expected syncedBlocks empty")
		}
	}()
	var tentativeState state.VirtualState
	if sm.solidState != nil {
		tentativeState = sm.solidState.Clone()
	} else {
		tentativeState = state.NewEmptyVirtualState(sm.chain.ID())
	}
	syncedBlocks := make([]state.Block, 0)
	for i := fromIndex; i < sm.stateOutput.GetStateIndex(); i++ {
		sb := sm.syncBlock(i)
		syncedBlocks = append(syncedBlocks, sb)
		if err := tentativeState.ApplyBlock(sb); err != nil {
			sm.log.Errorf("failed to apply synced block. Start syncing from scratch from block #%d to #%d. Error: %v",
				fromIndex, sm.stateOutput.GetStateIndex(), err)
			sm.syncedBlocks = make(map[uint32]*syncingBlock)
			return
		}
	}
	// state hashes must be equal
	stateHash1 := tentativeState.Hash()
	stateHash2, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		sm.log.Panicf("failed to decode state hash")
	}
	if stateHash1 != stateHash2 {
		sm.log.Errorf("state hashes mismatch between state and anchor transaction. Start syncing from scratch from block #%d to #%d",
			fromIndex, sm.stateOutput.GetStateIndex())
		sm.syncedBlocks = make(map[uint32]*syncingBlock)
		return
	}
	// again applying blocks, this time seriously
	if sm.solidState == nil {
		sm.solidState = state.NewEmptyVirtualState(sm.chain.ID())
	}
	for _, block := range syncedBlocks {
		if err := tentativeState.ApplyBlock(block); err != nil {
			sm.log.Errorf("inconsistency: %v", err)
			return
		}
		if err := sm.solidState.CommitToDb(block); err == nil {
			delete(sm.syncedBlocks, block.StateIndex())
		} else {
			sm.log.Errorf("failed to commit synced changes into DB. Restart syncing")
			sm.syncedBlocks = make(map[uint32]*syncingBlock)
			return
		}
	}
}
