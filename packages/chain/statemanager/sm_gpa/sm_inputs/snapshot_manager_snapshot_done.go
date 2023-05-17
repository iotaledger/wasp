package sm_inputs

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type SnapshotManagerSnapshotDone struct {
	stateIndex uint32
	commitment *state.L1Commitment
	result     error
}

var _ gpa.Input = &SnapshotManagerSnapshotDone{}

func NewSnapshotManagerSnapshotDone(stateIndex uint32, commitment *state.L1Commitment, result error) *SnapshotManagerSnapshotDone {
	return &SnapshotManagerSnapshotDone{
		stateIndex: stateIndex,
		commitment: commitment,
		result:     result,
	}
}

func (smsdT *SnapshotManagerSnapshotDone) GetStateIndex() uint32 {
	return smsdT.stateIndex
}

func (smsdT *SnapshotManagerSnapshotDone) GetCommitment() *state.L1Commitment {
	return smsdT.commitment
}

func (smsdT *SnapshotManagerSnapshotDone) GetResult() error {
	return smsdT.result
}
