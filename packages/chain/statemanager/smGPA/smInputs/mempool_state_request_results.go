package smInputs

import (
	"github.com/iotaledger/wasp/packages/state"
)

type MempoolStateRequestResults struct { // TODO: move to mempool package
	newState state.State   // state for newL1Commitment
	added    []state.Block // blocks from common to newL1Commitment (excluding common)
	removed  []state.Block // blocks from common to oldL1Commitment (excluding common)
}

func NewMempoolStateRequestResults(newState state.State, added, removed []state.Block) *MempoolStateRequestResults {
	return &MempoolStateRequestResults{
		newState: newState,
		added:    added,
		removed:  removed,
	}
}

func (msrrT *MempoolStateRequestResults) GetNewState() state.State {
	return msrrT.newState
}

func (msrrT *MempoolStateRequestResults) GetAdded() []state.Block {
	return msrrT.added
}

func (msrrT *MempoolStateRequestResults) GetRemoved() []state.Block {
	return msrrT.removed
}
