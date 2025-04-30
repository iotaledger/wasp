package inputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
)

type ConsensusDecidedState struct {
	context      context.Context
	stateIndex   uint32
	l1Commitment *state.L1Commitment
	resultCh     chan<- state.State
}

var _ gpa.Input = &ConsensusDecidedState{}

func NewConsensusDecidedState(ctx context.Context, anchor *isc.StateAnchor) (*ConsensusDecidedState, <-chan state.State) {
	commitment, err := transaction.L1CommitmentFromAnchor(anchor)
	if err != nil {
		panic("Cannot make L1 commitment from anchor")
	}
	resultChannel := make(chan state.State, 1)
	return &ConsensusDecidedState{
		context:      ctx,
		stateIndex:   anchor.GetStateIndex(),
		l1Commitment: commitment,
		resultCh:     resultChannel,
	}, resultChannel
}

func (cdsT *ConsensusDecidedState) GetStateIndex() uint32 {
	return cdsT.stateIndex
}

func (cdsT *ConsensusDecidedState) GetL1Commitment() *state.L1Commitment {
	return cdsT.l1Commitment
}

func (cdsT *ConsensusDecidedState) IsValid() bool {
	return cdsT.context.Err() == nil
}

func (cdsT *ConsensusDecidedState) Respond(theState state.State) {
	if cdsT.IsValid() && !cdsT.IsResultChClosed() {
		cdsT.resultCh <- theState
		cdsT.closeResultCh()
	}
}

func (cdsT *ConsensusDecidedState) IsResultChClosed() bool {
	return cdsT.resultCh == nil
}

func (cdsT *ConsensusDecidedState) closeResultCh() {
	close(cdsT.resultCh)
	cdsT.resultCh = nil
}
