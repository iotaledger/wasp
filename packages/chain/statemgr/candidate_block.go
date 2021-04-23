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
	local             bool
	votes             int
	approvingOutputID *ledgerstate.OutputID // TODO: block should have approvingOutputID; change it to `approved bool`
	stateHash         *hashing.HashValue    // TODO: may the pointer be removed?
}

func newCandidateBlock(block state.Block, stateHash *hashing.HashValue) *candidateBlock {
	var local bool
	if stateHash == nil {
		local = false
	} else {
		local = true
	}
	return &candidateBlock{
		block:             block,
		local:             local,
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

func (cThis *candidateBlock) isLocal() bool {
	return cThis.local
}

func (cThis *candidateBlock) isApproved() bool {
	return cThis.approvingOutputID != nil
}

func (cThis *candidateBlock) approveIfRightOutput(output *ledgerstate.AliasOutput) {
	if cThis.block.StateIndex() == output.GetStateIndex() {
		outputID := output.ID()
		finalHash, err := hashing.HashValueFromBytes(output.GetStateData())
		if err != nil {
			return
		}
		if cThis.isLocal() {
			if *(cThis.stateHash) == finalHash {
				cThis.approvingOutputID = &outputID
				cThis.block.WithApprovingOutputID(outputID)
			}
		} else {
			if cThis.block.ApprovingOutputID() == outputID {
				cThis.approvingOutputID = &outputID
				cThis.stateHash = &finalHash
			}
		}
	}
}

func (cThis *candidateBlock) getStateHash() hashing.HashValue {
	/*if cThis.stateHash == nil {
		return hashing.HashValue{}
	}*/
	return *(cThis.stateHash)
}

func (cThis *candidateBlock) getApprovindOutputID() ledgerstate.OutputID {
	/*if cThis.approvingOutputID == nil {
		return ledgerstate.OutputID{}
	}*/
	return *(cThis.approvingOutputID)
}
