// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
)

type candidateBlock struct {
	block         state.Block
	local         bool
	votes         int
	approved      bool
	nextStateHash hashing.HashValue
	nextState     state.VirtualState
}

func newCandidateBlock(block state.Block, nextStateIfProvided state.VirtualState) *candidateBlock {
	var local bool
	var stateHash hashing.HashValue
	if nextStateIfProvided == nil {
		local = false
		stateHash = hashing.NilHash
	} else {
		local = true
		stateHash = nextStateIfProvided.Hash()
	}
	return &candidateBlock{
		block:         block,
		local:         local,
		votes:         1,
		approved:      false,
		nextStateHash: stateHash,
		nextState:     nextStateIfProvided,
	}
}

func (ct *candidateBlock) getBlock() state.Block {
	return ct.block
}

func (ct *candidateBlock) addVote() {
	ct.votes++
}

func (ct *candidateBlock) getVotes() int {
	return ct.votes
}

func (ct *candidateBlock) isLocal() bool {
	return ct.local
}

func (ct *candidateBlock) isApproved() bool {
	return ct.approved
}

func (ct *candidateBlock) approveIfRightOutput(output *ledgerstate.AliasOutput) {
	if ct.block.BlockIndex() == output.GetStateIndex() {
		outputID := output.ID()
		finalHash, err := hashing.HashValueFromBytes(output.GetStateData())
		if err != nil {
			return
		}
		if ct.isLocal() {
			if ct.nextStateHash == finalHash {
				ct.approved = true
				ct.block.SetApprovingOutputID(outputID)
			}
		} else {
			if ct.block.ApprovingOutputID() == outputID {
				ct.approved = true
				ct.nextStateHash = finalHash
			}
		}
	}
}

func (ct *candidateBlock) getNextStateHash() hashing.HashValue {
	return ct.nextStateHash
}

func (ct *candidateBlock) getNextState(currentState state.VirtualState) (state.VirtualState, error) {
	if ct.isLocal() {
		return ct.nextState, nil
	}
	err := currentState.ApplyBlock(ct.block)
	return currentState, err
}

func (ct *candidateBlock) getApprovingOutputID() ledgerstate.OutputID {
	return ct.block.ApprovingOutputID()
}
