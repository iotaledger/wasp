// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/state"
)

type candidateBlock struct {
	block     state.Block
	nextState state.VirtualStateAccess
}

func newCandidateBlock(block state.Block, nextStateIfProvided state.VirtualStateAccess) *candidateBlock {
	return &candidateBlock{
		block:     block,
		nextState: nextStateIfProvided,
	}
}

func (cT *candidateBlock) getBlock() state.Block {
	return cT.block
}

func (cT *candidateBlock) getNextState(currentState state.VirtualStateAccess) (state.VirtualStateAccess, error) {
	if cT.nextState == nil {
		err := currentState.ApplyBlock(cT.block)
		return currentState, err
	}
	return cT.nextState.Copy(), nil
}

func (cT *candidateBlock) getApprovingOutputID() *iotago.UTXOInput {
	return cT.block.ApprovingOutputID()
}

func (cT *candidateBlock) getPreviousL1Commitment() *state.L1Commitment {
	return cT.block.PreviousL1Commitment()
}
