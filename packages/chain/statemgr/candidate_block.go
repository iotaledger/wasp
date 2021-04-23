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
	block     state.Block
	local     bool
	votes     int
	approved  bool
	stateHash hashing.HashValue
}

func newCandidateBlock(block state.Block, stateHashIfProvided *hashing.HashValue) *candidateBlock {
	var local bool
	var stateHash hashing.HashValue
	if stateHashIfProvided == nil {
		local = false
		stateHash = hashing.NilHash
	} else {
		local = true
		stateHash = *stateHashIfProvided
	}
	return &candidateBlock{
		block:     block,
		local:     local,
		votes:     1,
		approved:  false,
		stateHash: stateHash,
	}
}

func (cT *candidateBlock) getBlock() state.Block {
	return cT.block
}

func (cT *candidateBlock) addVote() {
	cT.votes++
}

func (cT *candidateBlock) getVotes() int {
	return cT.votes
}

func (cT *candidateBlock) isLocal() bool {
	return cT.local
}

func (cT *candidateBlock) isApproved() bool {
	return cT.approved
}

func (cT *candidateBlock) approveIfRightOutput(output *ledgerstate.AliasOutput) {
	if cT.block.StateIndex() == output.GetStateIndex() {
		outputID := output.ID()
		finalHash, err := hashing.HashValueFromBytes(output.GetStateData())
		if err != nil {
			return
		}
		if cT.isLocal() {
			if cT.stateHash == finalHash {
				cT.approved = true
				cT.block.WithApprovingOutputID(outputID)
			}
		} else {
			if cT.block.ApprovingOutputID() == outputID {
				cT.approved = true
				cT.stateHash = finalHash
			}
		}
	}
}

func (cT *candidateBlock) getStateHash() hashing.HashValue {
	return cT.stateHash
}

func (cT *candidateBlock) getApprovingOutputID() ledgerstate.OutputID {
	return cT.block.ApprovingOutputID()
}
