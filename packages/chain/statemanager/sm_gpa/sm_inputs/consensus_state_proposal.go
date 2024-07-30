package sm_inputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/types"
)

type ConsensusStateProposal struct {
	context      context.Context
	stateIndex   uint32
	l1Commitment *state.L1Commitment
	resultCh     chan<- interface{}
}

var _ gpa.Input = &ConsensusStateProposal{}

func NewConsensusStateProposal(ctx context.Context, anchor *types.RefWithObject[types.Anchor]) (*ConsensusStateProposal, <-chan interface{}) {
	commitment, err := state.NewL1CommitmentFromAnchor(anchor.Object)
	if err != nil {
		panic("Cannot make L1 commitment from anchor")
	}
	resultChannel := make(chan interface{}, 1)
	return &ConsensusStateProposal{
		context:      ctx,
		stateIndex:   anchor.Object.StateIndex,
		l1Commitment: commitment,
		resultCh:     resultChannel,
	}, resultChannel
}

func (cspT *ConsensusStateProposal) GetStateIndex() uint32 {
	return cspT.stateIndex
}

func (cspT *ConsensusStateProposal) GetL1Commitment() *state.L1Commitment {
	return cspT.l1Commitment
}

func (cspT *ConsensusStateProposal) IsValid() bool {
	return cspT.context.Err() == nil
}

func (cspT *ConsensusStateProposal) Respond() {
	if cspT.IsValid() && !cspT.IsResultChClosed() {
		cspT.resultCh <- nil
		cspT.closeResultCh()
	}
}

func (cspT *ConsensusStateProposal) IsResultChClosed() bool {
	return cspT.resultCh == nil
}

func (cspT *ConsensusStateProposal) closeResultCh() {
	close(cspT.resultCh)
	cspT.resultCh = nil
}
