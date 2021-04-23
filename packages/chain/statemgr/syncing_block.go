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
	pullDeadline    time.Time
	blockCandidates map[hashing.HashValue]*candidateBlock
}

func newSyncingBlocks(log *logger.Logger) *syncingBlocks {
	return &syncingBlocks{
		blocks: make(map[uint32]*syncingBlock),
		log:    log,
	}
}

func (syncsThis *syncingBlocks) getPullDeadline(stateIndex uint32) time.Time {
	sync, ok := syncsThis.blocks[stateIndex]
	if !ok {
		return time.Time{}
	}
	return sync.pullDeadline
}

func (syncsThis *syncingBlocks) setPullDeadline(stateIndex uint32, pullDeadline time.Time) {
	sync, ok := syncsThis.blocks[stateIndex]
	if ok {
		sync.pullDeadline = pullDeadline
	}
}

func (syncsThis *syncingBlocks) getBlockCandidates(stateIndex uint32) []*candidateBlock {
	sync, ok := syncsThis.blocks[stateIndex]
	if !ok {
		return make([]*candidateBlock, 0, 0)
	}
	result := make([]*candidateBlock, 0, len(sync.blockCandidates))
	for _, candidate := range sync.blockCandidates {
		result = append(result, candidate)
	}
	return result
}

func (syncsThis *syncingBlocks) getApprovedBlockCandidates(stateIndex uint32) []*candidateBlock {
	result := make([]*candidateBlock, 0, 1)
	sync, ok := syncsThis.blocks[stateIndex]
	if ok {
		for _, candidate := range sync.blockCandidates {
			if candidate.isApproved() {
				result = append(result, candidate)
			}
		}
	}
	return result
}

func (syncsThis *syncingBlocks) getBlockCandidatesCount(stateIndex uint32) int {
	sync, ok := syncsThis.blocks[stateIndex]
	if !ok {
		return 0
	}
	return len(sync.blockCandidates)
}

func (syncsThis *syncingBlocks) getApprovedBlockCandidatesCount(stateIndex uint32) int {
	sync, ok := syncsThis.blocks[stateIndex]
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

func (syncsThis *syncingBlocks) hasBlockCandidates() bool {
	for _, sync := range syncsThis.blocks {
		if len(sync.blockCandidates) > 0 {
			return true
		}
	}
	return false
}

func (syncsThis *syncingBlocks) addBlockCandidate(block state.Block, stateHash *hashing.HashValue) (isBlockNew bool, candidate *candidateBlock, err error) {
	stateIndex := block.StateIndex()
	syncsThis.startSyncingIfNeeded(stateIndex)
	sync, _ := syncsThis.blocks[stateIndex]
	hash := block.EssenceHash()
	candidateExisting, ok := sync.blockCandidates[hash]
	if ok {
		// already have block. Check consistency. If inconsistent, start from scratch
		if candidateExisting.getApprovingOutputID() != block.ApprovingOutputID() {
			delete(sync.blockCandidates, hash)
			return false, nil, fmt.Errorf("conflicting block arrived. Block index: %d, present approving outputID: %s, arrived approving outputID: %s",
				stateIndex, coretypes.OID(candidateExisting.getApprovingOutputID()), coretypes.OID(block.ApprovingOutputID()))

		}
		candidateExisting.addVote()
		syncsThis.log.Infof("added existing block candidate. State index: %d, state hash: %s", stateIndex, hash.String())
		return false, candidateExisting, nil
	}
	candidate = newCandidateBlock(block, stateHash)
	sync.blockCandidates[hash] = candidate
	sync.pullDeadline = time.Now().Add(periodBetweenSyncMessages * 2)
	syncsThis.log.Infof("added new block candidate. State index: %d, state hash: %s", stateIndex, hash.String())
	return true, candidate, nil
}

func (syncsThis *syncingBlocks) approveBlockCandidates(output *ledgerstate.AliasOutput) {
	syncsThis.log.Infof("XXX approveBlockCandidates %v", coretypes.OID(output.ID()))
	if output == nil {
		return
	}
	stateIndex := output.GetStateIndex()
	sync, ok := syncsThis.blocks[stateIndex]
	syncsThis.log.Infof("XXX approveBlockCandidates: candidates=%v", syncsThis.hasBlockCandidates())
	if ok {
		syncsThis.log.Infof("XXX approveBlockCandidates: sync block %v found", stateIndex)
		for i, candidate := range sync.blockCandidates {
			//syncsThis.log.Infof("XXX approveBlockCandidates: candidate %v local %v, approved %v, block hash %v output hash %v, block id %v output id %v", i, candidate.isLocal(), candidate.isApproved(), candidate.getStateHash(), finalHash, candidate.getApprovingOutputID(), outputID)
			//syncsThis.log.Infof("XXX approveBlockCandidates: candidate %v local %v, approved %v, output hash %v, output id %v", i, candidate.isLocal(), candidate.isApproved(), finalHash, outputID)
			syncsThis.log.Infof("XXX approveBlockCandidates: candidate %v local %v, approved %v", i, candidate.isLocal(), candidate.isApproved())
			candidate.approveIfRightOutput(output)
		}
	}
}

func (syncsThis *syncingBlocks) startSyncingIfNeeded(stateIndex uint32) {
	if !syncsThis.isSyncing(stateIndex) {
		syncsThis.blocks[stateIndex] = &syncingBlock{
			//pullDeadline      time.Time       // // TODO:
			blockCandidates: make(map[hashing.HashValue]*candidateBlock),
		}
	}
}

func (syncsThis *syncingBlocks) isSyncing(stateIndex uint32) bool {
	_, ok := syncsThis.blocks[stateIndex]
	return ok
}

func (syncsThis *syncingBlocks) restartSyncing() {
	syncsThis.blocks = make(map[uint32]*syncingBlock)
}

func (syncsThis *syncingBlocks) deleteSyncingBlock(stateIndex uint32) {
	delete(syncsThis.blocks, stateIndex)
}
