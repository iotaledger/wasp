package smInputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

type ConsensusDecidedState struct {
	context      context.Context
	l1Commitment *state.L1Commitment
	resultCh     chan<- state.State
}

var _ gpa.Input = &ConsensusDecidedState{}

func NewConsensusDecidedState(ctx context.Context, aliasOutput *isc.AliasOutputWithID) (*ConsensusDecidedState, <-chan state.State) {
	commitment, err := vmcontext.L1CommitmentFromAliasOutput(aliasOutput.GetAliasOutput())
	if err != nil {
		panic("Cannot make L1 commitment from alias output")
	}
	resultChannel := make(chan state.State, 1)
	return &ConsensusDecidedState{
		context:      ctx,
		l1Commitment: commitment,
		resultCh:     resultChannel,
	}, resultChannel
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
