// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
)

type syncingBlocks struct {
	blocks            map[uint32]*syncingBlock // StateIndex -> BlockCandidates
	log               *logger.Logger
	initialBlockRetry time.Duration
}

type syncingBlock struct {
	requestBlockRetryTime time.Time
	blockCandidates       map[hashing.HashValue]*candidateBlock
}

func newSyncingBlocks(log *logger.Logger, initialBlockRetry time.Duration) *syncingBlocks {
	return &syncingBlocks{
		blocks:            make(map[uint32]*syncingBlock),
		log:               log,
		initialBlockRetry: initialBlockRetry,
	}
}

func (syncsT *syncingBlocks) getRequestBlockRetryTime(stateIndex uint32) time.Time {
	if sync, ok := syncsT.blocks[stateIndex]; ok {
		return sync.requestBlockRetryTime
	}
	return time.Time{}
}

func (syncsT *syncingBlocks) setRequestBlockRetryTime(stateIndex uint32, requestBlockRetryTime time.Time) {
	if sync, ok := syncsT.blocks[stateIndex]; ok {
		sync.requestBlockRetryTime = requestBlockRetryTime
	}
}

func (syncsT *syncingBlocks) getBlockCandidates(stateIndex uint32) []*candidateBlock {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return make([]*candidateBlock, 0)
	}
	result := make([]*candidateBlock, len(sync.blockCandidates))
	i := 0
	for _, candidate := range sync.blockCandidates {
		result[i] = candidate
		i++
	}
	return result
}

func (syncsT *syncingBlocks) getApprovedBlockCandidates(stateIndex uint32) []*candidateBlock {
	result := make([]*candidateBlock, 0, 1)
	sync, ok := syncsT.blocks[stateIndex]
	if ok {
		for _, candidate := range sync.blockCandidates {
			if candidate.isApproved() {
				result = append(result, candidate)
			}
		}
	}
	return result
}

func (syncsT *syncingBlocks) getBlockCandidatesCount(stateIndex uint32) int {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return 0
	}
	return len(sync.blockCandidates)
}

func (syncsT *syncingBlocks) getApprovedBlockCandidatesCount(stateIndex uint32) int {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return 0
	}
	approvedCount := 0
	for _, candidate := range sync.blockCandidates {
		if candidate.isApproved() {
			approvedCount++
		}
	}
	return approvedCount
}

func (syncsT *syncingBlocks) hasBlockCandidates() bool {
	return syncsT.hasBlockCandidatesNotOlderThan(0)
}

func (syncsT *syncingBlocks) hasBlockCandidatesNotOlderThan(index uint32) bool {
	for i, sync := range syncsT.blocks {
		if i >= index {
			if len(sync.blockCandidates) > 0 {
				return true
			}
		}
	}
	return false
}

func (syncsT *syncingBlocks) addBlockCandidate(block state.Block, nextState state.VirtualState) (isBlockNew bool, candidate *candidateBlock) {
	stateIndex := block.BlockIndex()
	hash := hashing.HashData(block.EssenceBytes())
	syncsT.log.Debugf("addBlockCandidate: adding block candidate for index %v with essence hash %v; next state provided: %v", stateIndex, hash.String(), nextState != nil)
	syncsT.startSyncingIfNeeded(stateIndex)
	sync := syncsT.blocks[stateIndex]
	candidateExisting, ok := sync.blockCandidates[hash]
	if ok {
		// already have block. Check consistency. If inconsistent, start from scratch
		if candidateExisting.getApprovingOutputID() != block.ApprovingOutputID() {
			delete(sync.blockCandidates, hash)
			syncsT.log.Debugf("addBlockCandidate: conflicting block index %v with hash %v arrived: prsent approvingOutputID %v, new block approvingOutputID: %v",
				stateIndex, hash.String(), candidateExisting.getApprovingOutputID(), iscp.OID(block.ApprovingOutputID()))
			return false, nil
		}
		candidateExisting.addVote()
		syncsT.log.Debugf("addBlockCandidate: existing block index %v with hash %v arrived, votes increased.", stateIndex, hash.String())
		return false, candidateExisting
	}
	candidate = newCandidateBlock(block, nextState)
	sync.blockCandidates[hash] = candidate
	syncsT.log.Debugf("addBlockCandidate: new block candidate created for block index: %d, hash: %s", stateIndex, hash.String())
	return true, candidate
}

func (syncsT *syncingBlocks) approveBlockCandidates(output *ledgerstate.AliasOutput) bool {
	if output == nil {
		syncsT.log.Debugf("approveBlockCandidates failed, provided output is nil")
		return false
	}
	someApproved := false
	stateIndex := output.GetStateIndex()
	syncsT.log.Debugf("approveBlockCandidates using output ID %v for state index %v", iscp.OID(output.ID()), stateIndex)
	sync, ok := syncsT.blocks[stateIndex]
	if ok {
		syncsT.log.Debugf("approveBlockCandidates: %v block candidates to check", len(sync.blockCandidates))
		for blockHash, candidate := range sync.blockCandidates {
			alreadyApproved := candidate.isApproved()
			syncsT.log.Debugf("approveBlockCandidates: checking candidate %v: local %v, nextStateHash %v, approvingOutputID %v, already approved %v",
				blockHash.String(), candidate.isLocal(), candidate.getNextStateHash().String(), iscp.OID(candidate.getApprovingOutputID()), alreadyApproved)
			if !alreadyApproved {
				candidate.approveIfRightOutput(output)
				if candidate.isApproved() {
					syncsT.log.Debugf("approveBlockCandidates: candidate %v got approved", blockHash.String())
					someApproved = true
				}
			}
		}
	}
	return someApproved
}

func (syncsT *syncingBlocks) startSyncingIfNeeded(stateIndex uint32) {
	if !syncsT.isSyncing(stateIndex) {
		syncsT.log.Debugf("Starting syncing state index %v", stateIndex)
		syncsT.blocks[stateIndex] = &syncingBlock{
			requestBlockRetryTime: time.Now().Add(syncsT.initialBlockRetry),
			blockCandidates:       make(map[hashing.HashValue]*candidateBlock),
		}
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
