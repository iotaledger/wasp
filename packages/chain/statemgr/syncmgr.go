package statemgr

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

// returns block if it is known and flag if it is approved by some output
func (sm *stateManager) syncBlock(blockIndex uint32) (state.Block, bool) {
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
}

func (sm *stateManager) blockArrived(block state.Block) {
	syncBlk, ok := sm.syncingBlocks[block.BlockIndex()]
	if !ok {
		// not asked
		return
	}
	if syncBlk.block != nil {
		// already have block. Check consistency. If inconsistent, start from scratch
		if syncBlk.block.ApprovingOutputID() != block.ApprovingOutputID() {
			sm.log.Errorf("conflicting block arrived. Block index: %d, present approving outputID: %s, arrived approving outputID: %s",
				block.BlockIndex(), coretypes.OID(syncBlk.block.ApprovingOutputID()), coretypes.OID(block.ApprovingOutputID()))
			syncBlk.block = nil
			return
		}
		if !bytes.Equal(syncBlk.block.EssenceBytes(), block.EssenceBytes()) {
			sm.log.Errorf("conflicting block arrived. Block index: %d", block.BlockIndex())
			syncBlk.block = nil
			return
		}
		return
	}
	// new block
	// ask for approving output
	sm.nodeConn.PullConfirmedOutput(sm.chain.ID().AsAddress(), block.ApprovingOutputID())
	syncBlk.block = block
	syncBlk.pullDeadline = time.Now().Add(periodBetweenSyncMessages * 2)
}

func (sm *stateManager) chainOutputArrived(chainOutput *ledgerstate.AliasOutput) {
	syncBlk, ok := sm.syncingBlocks[chainOutput.GetStateIndex()]
	if !ok {
		// not interested
		return
	}
	if syncBlk.block == nil || syncBlk.approved {
		// no need yet or too late
		return
	}
	if BlockIsApprovedByChainOutput(syncBlk.block, chainOutput) {
		finalHash, err := hashing.HashValueFromBytes(chainOutput.GetStateData())
		if err != nil {
			return
		}
		syncBlk.finalHash = finalHash
		syncBlk.approvingOutputID = chainOutput.ID()
		syncBlk.approved = true
	}
}

func BlockIsApprovedByChainOutput(b state.Block, chainOutput *ledgerstate.AliasOutput) bool {
	if chainOutput == nil {
		return false
	}
	if b.BlockIndex() != chainOutput.GetStateIndex() {
		return false
	}
	var nilOID ledgerstate.OutputID
	if b.ApprovingOutputID() != nilOID && b.ApprovingOutputID() != chainOutput.ID() {
		return false
	}
	return true
}

func (sm *stateManager) doSyncActionIfNeeded() {
	if sm.stateOutput == nil {
		return
	}
	currentIndex := sm.solidState.BlockIndex()

	switch {
	case currentIndex == sm.stateOutput.GetStateIndex():
		// synced
		return
	case currentIndex > sm.stateOutput.GetStateIndex():
		sm.log.Panicf("inconsistency: current (solid) state index %d is larger than state output index %d",
			currentIndex, sm.stateOutput.GetStateIndex())
	}
	// not synced
	if currentIndex+1 >= sm.stateOutput.GetStateIndex() {
		// waiting for the state transition
		return
	}
	for i := currentIndex + 1; i < sm.stateOutput.GetStateIndex(); i++ {
		block, approved := sm.syncBlock(i)
		if block == nil {
			// some block are still unknown. Can't sync
			return
		}
		if approved {
			blocks := make([]state.Block, 0)
			for j := currentIndex + 1; j <= i; j++ {
				b, _ := sm.syncBlock(j)
				blocks = append(blocks, b)
			}
			sm.mustCommitSynced(blocks, sm.syncingBlocks[i].finalHash, sm.syncingBlocks[i].approvingOutputID)
			return
		}
	}
	// nothing is approved but all blocks are here
	blocks := make([]state.Block, 0)
	for i := currentIndex + 1; i < sm.stateOutput.GetStateIndex(); i++ {
		blocks = append(blocks, sm.syncingBlocks[i].block)
	}
	finalHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		return
	}
	sm.mustCommitSynced(blocks, finalHash, sm.stateOutput.ID())
}

// assumes all synced already
func (sm *stateManager) mustCommitSynced(blocks []state.Block, finalHash hashing.HashValue, outputID ledgerstate.OutputID) {
	if len(blocks) == 0 {
		// shouldn't be here
		sm.log.Panicf("len(blocks) == 0")
	}
	tentativeState := sm.solidState.Clone()
	for _, block := range blocks {
		if err := tentativeState.ApplyBlock(block); err != nil {
			sm.log.Errorf("failed to apply synced block index #%d. Error: %v", block.BlockIndex(), err)
			return
		}
	}
	// state hashes must be equal
	tentativeHash := tentativeState.Hash()
	if tentativeHash != finalHash {
		sm.log.Errorf("state hashes mismatch: expected final hash: %s, tentative hash: %s", finalHash, tentativeHash)
		return
	}
	tentativeState = nil
	// running it again, this time seriously
	// the reason we are not committing all at once is to prevent too large DB transaction
	stateIndex := uint32(0)
	for _, block := range blocks {
		stateIndex = block.BlockIndex()
		if err := sm.solidState.ApplyBlock(block); err != nil {
			sm.log.Errorf("failed to apply synced block index #%d. Error: %v", stateIndex, err)
			return
		}
		// one block at a time
		if err := sm.solidState.Commit(block); err != nil {
			sm.log.Errorf("failed to commit synced changes into DB. Restart syncing")
			sm.syncingBlocks = make(map[uint32]*syncingBlock)
			return
		}
		delete(sm.syncingBlocks, stateIndex)
	}
	go sm.chain.Events().StateSynced().Trigger(outputID, stateIndex)
}
