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

func (cThis *candidateBlock) getBlock() state.Block {
	return cThis.block
}

func (cThis *candidateBlock) addVote() {
	cThis.votes++
}

func (cThis *candidateBlock) getVotes() int {
	return cThis.votes
}

func (cThis *candidateBlock) isLocal() bool {
	return cThis.local
}

func (cThis *candidateBlock) isApproved() bool {
	return cThis.approved
}

func (cThis *candidateBlock) approveIfRightOutput(output *ledgerstate.AliasOutput) {
	if cThis.block.StateIndex() == output.GetStateIndex() {
		outputID := output.ID()
		finalHash, err := hashing.HashValueFromBytes(output.GetStateData())
		if err != nil {
			return
		}
		if cThis.isLocal() {
			if cThis.stateHash == finalHash {
				cThis.approved = true
				cThis.block.WithApprovingOutputID(outputID)
			}
		} else {
			if cThis.block.ApprovingOutputID() == outputID {
				cThis.approved = true
				cThis.stateHash = finalHash
			}
		}
	}
}

func (cThis *candidateBlock) getStateHash() hashing.HashValue {
	return cThis.stateHash
}

func (cThis *candidateBlock) getApprovingOutputID() ledgerstate.OutputID {
	return cThis.block.ApprovingOutputID()
}
