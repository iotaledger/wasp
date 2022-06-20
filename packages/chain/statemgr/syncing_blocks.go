// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
)

const (
	pollFallbackDelay = 5 * time.Second
)

type syncingBlocks struct {
	blocks       map[uint32]*syncingBlock // StateIndex -> BlockCandidates
	log          *logger.Logger
	wal          chain.WAL
	lastPullTime time.Time // Time, when we pulled for some block last time.
	lastRecvTime time.Time // Time, when we received any block we pulled. Used to determine, if fallback nodes should be used.
}

func newSyncingBlocks(log *logger.Logger, wal chain.WAL) *syncingBlocks {
	return &syncingBlocks{
		blocks: make(map[uint32]*syncingBlock),
		log:    log,
		wal:    wal,
	}
}

func (syncsT *syncingBlocks) getRequestBlockRetryTime(stateIndex uint32) time.Time {
	if sync, ok := syncsT.blocks[stateIndex]; ok {
		return sync.getRequestBlockRetryTime()
	}
	return time.Time{}
}

func (syncsT *syncingBlocks) setRequestBlockRetryTime(stateIndex uint32, requestBlockRetryTime time.Time) {
	if sync, ok := syncsT.blocks[stateIndex]; ok {
		sync.setRequestBlockRetryTime(requestBlockRetryTime)
	}
}

func (syncsT *syncingBlocks) getBlockCandidatesCount(stateIndex uint32) int {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return 0
	}
	return sync.getBlockCandidatesCount()
}

func (syncsT *syncingBlocks) getBlockCandidate(stateIndex uint32, hash state.BlockHash) *candidateBlock {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return nil
	}
	return sync.getBlockCandidate(hash)
}

func (syncsT *syncingBlocks) hasApprovedBlockCandidate(stateIndex uint32) bool {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return false
	}
	return sync.hasApprovedBlockCandidate()
}

func (syncsT *syncingBlocks) getApprovedBlockCandidateHash(stateIndex uint32) state.BlockHash {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return state.BlockHash{}
	}
	return sync.getApprovedBlockCandidateHash()
}

func (syncsT *syncingBlocks) getNextStateCommitment(stateIndex uint32) trie.VCommitment {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return nil
	}
	return sync.getNextStateCommitment()
}

func (syncsT *syncingBlocks) hasBlockCandidates() bool {
	return syncsT.hasBlockCandidatesNotOlderThan(0)
}

func (syncsT *syncingBlocks) hasBlockCandidatesNotOlderThan(index uint32) bool {
	for i, sync := range syncsT.blocks {
		if i >= index {
			if sync.getBlockCandidatesCount() > 0 {
				return true
			}
		}
	}
	return false
}

func (syncsT *syncingBlocks) addBlockCandidate(block state.Block, nextState state.VirtualStateAccess) {
	stateIndex := block.BlockIndex()
	hash := state.BlockHashFromData(block.EssenceBytes())
	syncsT.log.Debugf("addBlockCandidate: adding block candidate for index %v with essence hash %s; next state provided: %v", stateIndex, hash, nextState != nil)
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		syncsT.log.Errorf("addBlockCandidate: adding block candidate for index %v with essence hash %s failed: index is not syncing", stateIndex, hash)
		return
	}
	sync.addBlockCandidate(hash, block, nextState)
}

func (syncsT *syncingBlocks) setApprovalInfo(output *iscp.AliasOutputWithID) {
	if output == nil {
		syncsT.log.Debugf("setApprovalInfo failed, provided output is nil")
		return
	}
	stateIndex := output.GetStateIndex()
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		syncsT.log.Debugf("setApprovalInfo failed: state index %v is not syncing", stateIndex)
		return
	}
	sync.setApprovalInfo(output)
}

func (syncsT *syncingBlocks) isObtainedFromWAL(i uint32) bool {
	sync, ok := syncsT.blocks[i]
	if ok {
		return sync.isReceivedFromWAL()
	}
	return false
}

func (syncsT *syncingBlocks) startSyncingIfNeeded(stateIndex uint32) {
	if !syncsT.isSyncing(stateIndex) {
		syncsT.log.Debugf("startSyncingIfNeeded: starting syncing state index %v", stateIndex)
		syncsT.blocks[stateIndex] = newSyncingBlock(syncsT.log.Named(fmt.Sprint(stateIndex)))

		// Getting block from write ahead log, if available
		if !syncsT.wal.Contains(stateIndex) {
			syncsT.log.Debugf("startSyncingIfNeeded: block with index %d not found in wal.", stateIndex)
			return
		}
		blockBytes, err := syncsT.wal.Read(stateIndex)
		if err != nil {
			syncsT.log.Errorf("startSyncingIfNeeded: error reading block bytes for index %d from wal: %v", stateIndex, err)
			return
		}
		block, err := state.BlockFromBytes(blockBytes)
		if err != nil {
			syncsT.log.Errorf("startSyncingIfNeeded: error obtaining block from block bytes in wal for index %d: %v", stateIndex, err)
			return
		}
		syncsT.addBlockCandidate(block, nil)
		syncsT.blocks[stateIndex].setReceivedFromWAL()
		syncsT.log.Debugf("startSyncingIfNeeded: block with index %d included from wal.", stateIndex)
	}
}

func (syncsT *syncingBlocks) isSyncing(stateIndex uint32) bool {
	_, ok := syncsT.blocks[stateIndex]
	return ok
}

func (syncsT *syncingBlocks) restartSyncing() {
	syncsT.blocks = make(map[uint32]*syncingBlock)
}

func (syncsT *syncingBlocks) deleteSyncingBlock(stateIndex uint32) {
	delete(syncsT.blocks, stateIndex)
}

//
// Track poll/reception times, to determine, if no one responds to our polls.
//

func (syncsT *syncingBlocks) blocksPulled() {
	syncsT.lastPullTime = time.Now()
}

func (syncsT *syncingBlocks) blockReceived() {
	syncsT.lastRecvTime = time.Now()
}

func (syncsT *syncingBlocks) blockPollFallbackNeeded() bool {
	if len(syncsT.blocks) == 0 {
		return false
	}
	return syncsT.lastPullTime.Sub(syncsT.lastRecvTime) >= pollFallbackDelay
}
