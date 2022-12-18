package smInputs

import "github.com/iotaledger/wasp/packages/state"

type ChainFetchStateDiffResults struct {
	newState state.State   // state for newL1Commitment
	added    []state.Block // blocks from common to newL1Commitment (excluding common)
	removed  []state.Block // blocks from common to oldL1Commitment (excluding common)
}

func NewChainFetchStateDiffResults(newState state.State, added, removed []state.Block) *ChainFetchStateDiffResults {
	return &ChainFetchStateDiffResults{
		newState: newState,
		added:    added,
		removed:  removed,
	}
}

func (msrrT *ChainFetchStateDiffResults) GetNewState() state.State {
	return msrrT.newState
}

func (msrrT *ChainFetchStateDiffResults) GetAdded() []state.Block {
	return msrrT.added
}

func (msrrT *ChainFetchStateDiffResults) GetRemoved() []state.Block {
	return msrrT.removed
}
