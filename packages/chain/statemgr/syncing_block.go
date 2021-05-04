// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

type syncingBlocks struct {
	blocks map[uint32]*syncingBlock
	log    *logger.Logger
}

type syncingBlock struct {
	requestBlockRetryTime time.Time
	blockCandidates       map[hashing.HashValue]*candidateBlock
}

func newSyncingBlocks(log *logger.Logger) *syncingBlocks {
	return &syncingBlocks{
		blocks: make(map[uint32]*syncingBlock),
		log:    log,
	}
}

func (syncsT *syncingBlocks) getRequestBlockRetryTime(stateIndex uint32) time.Time {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return time.Time{}
	}
	return sync.requestBlockRetryTime
}

func (syncsT *syncingBlocks) setRequestBlockRetryTime(stateIndex uint32, requestBlockRetryTime time.Time) {
	sync, ok := syncsT.blocks[stateIndex]
	if ok {
		sync.requestBlockRetryTime = requestBlockRetryTime
	}
}

func (syncsT *syncingBlocks) getBlockCandidates(stateIndex uint32) []*candidateBlock {
	sync, ok := syncsT.blocks[stateIndex]
	if !ok {
		return make([]*candidateBlock, 0, 0)
	}
	result := make([]*candidateBlock, 0, len(sync.blockCandidates))
	for _, candidate := range sync.blockCandidates {
		result = append(result, candidate)
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

func (syncsT *syncingBlocks) addBlockCandidate(block state.Block, nextState state.VirtualState) (isBlockNew bool, candidate *candidateBlock, err error) {
	stateIndex := block.BlockIndex()
	syncsT.startSyncingIfNeeded(stateIndex)
	sync, _ := syncsT.blocks[stateIndex]
	hash := hashing.HashData(block.EssenceBytes())
	candidateExisting, ok := sync.blockCandidates[hash]
	if ok {
		// already have block. Check consistency. If inconsistent, start from scratch
		if candidateExisting.getApprovingOutputID() != block.ApprovingOutputID() {
			delete(sync.blockCandidates, hash)
			return false, nil, fmt.Errorf("conflicting block arrived. Block index: %d, present approving outputID: %s, arrived approving outputID: %s",
				stateIndex, coretypes.OID(candidateExisting.getApprovingOutputID()), coretypes.OID(block.ApprovingOutputID()))

		}
		candidateExisting.addVote()
		syncsT.log.Infof("added existing block candidate. State index: %d, state hash: %s", stateIndex, hash.String())
		return false, candidateExisting, nil
	}
	candidate = newCandidateBlock(block, nextState)
	sync.blockCandidates[hash] = candidate
	syncsT.log.Infof("added new block candidate. State index: %d, state hash: %s", stateIndex, hash.String())
	return true, candidate, nil
}

func (syncsT *syncingBlocks) approveBlockCandidates(output *ledgerstate.AliasOutput) {
	syncsT.log.Infof("XXX approveBlockCandidates %v", coretypes.OID(output.ID()))
	if output == nil {
		return
	}
	stateIndex := output.GetStateIndex()
	sync, ok := syncsT.blocks[stateIndex]
	syncsT.log.Infof("XXX approveBlockCandidates: candidates=%v", syncsT.hasBlockCandidates())
	if ok {
		syncsT.log.Infof("XXX approveBlockCandidates: sync block %v found", stateIndex)
		for i, candidate := range sync.blockCandidates {
			//syncsT.log.Infof("XXX approveBlockCandidates: candidate %v local %v, approved %v, block hash %v output hash %v, block id %v output id %v", i, candidate.isLocal(), candidate.isApproved(), candidate.getStateHash(), finalHash, candidate.getApprovingOutputID(), outputID)
			//syncsT.log.Infof("XXX approveBlockCandidates: candidate %v local %v, approved %v, output hash %v, output id %v", i, candidate.isLocal(), candidate.isApproved(), finalHash, outputID)
			syncsT.log.Infof("XXX approveBlockCandidates: candidate %v local %v, approved %v", i, candidate.isLocal(), candidate.isApproved())
			candidate.approveIfRightOutput(output)
		}
	}
}

func (syncsT *syncingBlocks) startSyncingIfNeeded(stateIndex uint32) {
	if !syncsT.isSyncing(stateIndex) {
		syncsT.blocks[stateIndex] = &syncingBlock{
			requestBlockRetryTime: time.Now(),
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
