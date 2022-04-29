// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/state"
)

type syncingBlock struct {
	requestBlockRetryTime time.Time
	blockCandidates       map[hashing.HashValue]*candidateBlock
	approvalInfo          *approvalInfo
	receivedFromWAL       bool
	log                   *logger.Logger
}

func newSyncingBlock(log *logger.Logger) *syncingBlock {
	return &syncingBlock{
		requestBlockRetryTime: time.Time{},
		blockCandidates:       make(map[hashing.HashValue]*candidateBlock),
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

func (syncT *syncingBlock) getBlockCandidate(hash hashing.HashValue) *candidateBlock {
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

func (syncT *syncingBlock) getApprovedBlockCandidateHash() hashing.HashValue {
	if syncT.approvalInfo == nil {
		return hashing.NilHash
	}
	return syncT.approvalInfo.getBlockHash()
}

func (syncT *syncingBlock) getNextStateCommitment() trie.VCommitment {
	if syncT.approvalInfo == nil {
		return nil
	}
	return syncT.approvalInfo.getNextStateCommitment()
}

func (syncT *syncingBlock) addBlockCandidate(hash hashing.HashValue, block state.Block, nextState state.VirtualStateAccess) (isBlockNew bool, candidate *candidateBlock) {
	candidateExisting, ok := syncT.blockCandidates[hash]
	if ok {
		// already have block. Check consistency. If inconsistent, start from scratch
		if !candidateExisting.getApprovingOutputID().Equals(block.ApprovingOutputID()) {
			delete(syncT.blockCandidates, hash)
			syncT.log.Warnf("addBlockCandidate: conflicting block index %v with hash %s arrived: present approvingOutputID %v, new block approvingOutputID: %v",
				block.BlockIndex(), hash, iscp.OID(candidateExisting.getApprovingOutputID()), iscp.OID(block.ApprovingOutputID()))
			return false, nil
		}
		candidateExisting.addVote()
		syncT.log.Debugf("addBlockCandidate: existing block index %v with hash %s arrived, votes increased.", block.BlockIndex(), hash)
		return false, candidateExisting
	}
	candidate = newCandidateBlock(block, nextState)
	syncT.blockCandidates[hash] = candidate
	syncT.log.Debugf("addBlockCandidate: new block candidate created for block index: %d, hash: %s", block.BlockIndex(), hash)
	return true, candidate
}

func (syncT *syncingBlock) setApprovalInfo(output *iscp.AliasOutputWithID) {
	approvalInfo, err := newApprovalInfo(output)
	if err != nil {
		syncT.log.Errorf("setApprovalInfo failed: %v", err)
		return
	}
	syncT.approvalInfo = approvalInfo
	syncT.log.Debugf("setApprovalInfo suceeded for state %v; approval info is %s", output.GetStateIndex(), approvalInfo)
}

func (syncT *syncingBlock) setReceivedFromWAL() {
	syncT.receivedFromWAL = true
}

func (syncT *syncingBlock) isReceivedFromWAL() bool {
	return syncT.receivedFromWAL
}
