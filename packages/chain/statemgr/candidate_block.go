// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

type candidateBlock struct {
	// block of state updates, not validated yet
	block             state.Block
	votes             int
	approvingOutputID *ledgerstate.OutputID
	stateHash         *hashing.HashValue
}

func newCandidateBlock(block state.Block, stateHash *hashing.HashValue) *candidateBlock {
	return &candidateBlock{
		block:             block,
		votes:             1,
		approvingOutputID: nil,
		stateHash:         stateHash,
	}
}

func (cThis *candidateBlock) getBlock() state.Block {
	return cThis.block
}

func (cThis *candidateBlock) addVote() {
	cThis.votes++
}

func (cThis *candidateBlock) isApproved() bool {
	return cThis.approvingOutputID != nil
}

func (cThis *candidateBlock) approveIfPossible(stateHash hashing.HashValue, approvingOutputID ledgerstate.OutputID) {
	if cThis.stateHash == nil {
		cThis.approvingOutputID = &approvingOutputID
		cThis.stateHash = &stateHash
	} else if *(cThis.stateHash) == stateHash {
		cThis.block.WithApprovingOutputID(approvingOutputID)
		cThis.approvingOutputID = &approvingOutputID
	}
}

func (cThis *candidateBlock) getStateHash() hashing.HashValue {
	if cThis.stateHash == nil {
		return hashing.HashValue{}
	}
	return *(cThis.stateHash)
}

func (cThis *candidateBlock) getApprovindOutputID() ledgerstate.OutputID {
	if cThis.approvingOutputID == nil {
		return ledgerstate.OutputID{}
	}
	return *(cThis.approvingOutputID)
}
