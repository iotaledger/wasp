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
		stateHash = nextStateIfProvided.StateCommitment()
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
	if cT.block.BlockIndex() == output.GetStateIndex() {
		outputID := output.ID()
		finalHash, err := hashing.HashValueFromBytes(output.GetStateData())
		if err != nil {
			return
		}
		if cT.isLocal() {
			if cT.nextStateHash == finalHash {
				cT.approved = true
				cT.block.SetApprovingOutputID(outputID)
			}
		} else {
			if cT.block.ApprovingOutputID() == outputID {
				cT.approved = true
				cT.nextStateHash = finalHash
			}
		}
	}
}

func (cT *candidateBlock) getNextStateHash() hashing.HashValue {
	return cT.nextStateHash
}

func (cT *candidateBlock) getNextState(currentState state.VirtualState) (state.VirtualState, error) {
	if cT.isLocal() {
		return cT.nextState, nil
	}
	err := currentState.ApplyBlock(cT.block)
	return currentState, err
}

func (cT *candidateBlock) getApprovingOutputID() ledgerstate.OutputID {
	return cT.block.ApprovingOutputID()
}
