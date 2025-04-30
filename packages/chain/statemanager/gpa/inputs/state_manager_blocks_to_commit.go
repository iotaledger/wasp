package inputs

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type StateManagerBlocksToCommit struct {
	commitments []*state.L1Commitment
}

var _ gpa.Input = &StateManagerBlocksToCommit{}

func NewStateManagerBlocksToCommit(commitments []*state.L1Commitment) *StateManagerBlocksToCommit {
	return &StateManagerBlocksToCommit{commitments: commitments}
}

func (smbtcT *StateManagerBlocksToCommit) GetCommitments() []*state.L1Commitment {
	return smbtcT.commitments
}
