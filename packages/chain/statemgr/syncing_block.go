// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type syncingBlock struct {
	requestBlockRetryTime time.Time
	blockCandidates       map[state.BlockHash]*candidateBlock
	approvalInfo          *approvalInfo
	receivedFromWAL       bool
	log                   *logger.Logger
}

func newSyncingBlock(log *logger.Logger) *syncingBlock {
	return &syncingBlock{
		requestBlockRetryTime: time.Time{},
		blockCandidates:       make(map[state.BlockHash]*candidateBlock),
		approvalInfo:          nil,
		receivedFromWAL:       false,
		log:                   log,
	}
}

func (syncT *syncingBlock) getRequestBlockRetryTime() time.Time {
	return syncT.requestBlockRetryTime
}

func (syncT *syncingBlock) setRequestBlockRetryTime(requestBlockRetryTime time.Time) {
	syncT.requestBlockRetryTime = requestBlockRetryTime
}

func (syncT *syncingBlock) getBlockCandidatesCount() int {
	return len(syncT.blockCandidates)
}

func (syncT *syncingBlock) getBlockCandidate(hash state.BlockHash) *candidateBlock {
	result, ok := syncT.blockCandidates[hash]
	if !ok {
		return nil
	}
	return result
}

func (syncT *syncingBlock) hasApprovedBlockCandidate() bool {
	if syncT.approvalInfo == nil {
		return false
	}
	return syncT.getBlockCandidate(syncT.approvalInfo.getBlockHash()) != nil
}

func (syncT *syncingBlock) getApprovedBlockCandidateHash() state.BlockHash {
	if syncT.approvalInfo == nil {
		return state.BlockHash{}
	}
	return syncT.approvalInfo.getBlockHash()
}

func (syncT *syncingBlock) getNextStateCommitment() common.VCommitment {
	if syncT.approvalInfo == nil {
		return nil
	}
	return syncT.approvalInfo.getNextStateCommitment()
}

func (syncT *syncingBlock) addBlockCandidate(hash state.BlockHash, block state.Block, nextState state.StateDraft) (isBlockNew bool, candidate *candidateBlock) {
	panic("TODO")
	candidateExisting, ok := syncT.blockCandidates[hash]
	if ok {
		// already have block. Check consistency. If inconsistent, start from scratch
		if !candidateExisting.getApprovingOutputID().Equals(block.ApprovingOutputID()) {
			delete(syncT.blockCandidates, hash)
			syncT.log.Warnf("addBlockCandidate: conflicting block index %v with hash %s arrived: present approvingOutputID %v, new block approvingOutputID: %v",
				nextState.BlockIndex(), hash, isc.OID(candidateExisting.getApprovingOutputID()), isc.OID(block.ApprovingOutputID()))
			return false, nil
		}
		syncT.log.Debugf("addBlockCandidate: existing block index %v with hash %s arrived, votes increased.", nextState.BlockIndex(), hash)
		return false, candidateExisting
	}
	candidate = newCandidateBlock(block, nextState)
	syncT.blockCandidates[hash] = candidate
	syncT.log.Debugf("addBlockCandidate: new block candidate created for block index: %d, hash: %s", nextState.BlockIndex(), hash)
	return true, candidate
}

func (syncT *syncingBlock) setApprovalInfo(output *isc.AliasOutputWithID) {
	approvalInfo, err := newApprovalInfo(output)
	if err != nil {
		syncT.log.Errorf("setApprovalInfo failed: %v", err)
		return
	}
	syncT.approvalInfo = approvalInfo
	syncT.log.Debugf("setApprovalInfo succeeded for state %v; approval info is %s", output.GetStateIndex(), approvalInfo)
}

func (syncT *syncingBlock) setReceivedFromWAL() {
	syncT.receivedFromWAL = true
}

func (syncT *syncingBlock) isReceivedFromWAL() bool {
	return syncT.receivedFromWAL
}
