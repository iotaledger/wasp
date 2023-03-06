package smInputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

type ChainFetchStateDiff struct {
	context         context.Context
	oldStateIndex   uint32
	newStateIndex   uint32
	oldL1Commitment *state.L1Commitment
	newL1Commitment *state.L1Commitment
	resultCh        chan<- *ChainFetchStateDiffResults
}

var _ gpa.Input = &ChainFetchStateDiff{}

func NewChainFetchStateDiff(ctx context.Context, prevAO, nextAO *isc.AliasOutputWithID) (*ChainFetchStateDiff, <-chan *ChainFetchStateDiffResults) {
	if prevAO == nil {
		// Only the current state is needed, if prevAO is unknown.
		prevAO = nextAO
	}
	oldCommitment, err := vmcontext.L1CommitmentFromAliasOutput(prevAO.GetAliasOutput())
	if err != nil {
		panic("Cannot make L1 commitment from previous alias output")
	}
	newCommitment, err := vmcontext.L1CommitmentFromAliasOutput(nextAO.GetAliasOutput())
	if err != nil {
		panic("Cannot make L1 commitment from next alias output")
	}
	resultChannel := make(chan *ChainFetchStateDiffResults, 1)
	return &ChainFetchStateDiff{
		context:         ctx,
		oldStateIndex:   prevAO.GetStateIndex(),
		newStateIndex:   nextAO.GetStateIndex(),
		oldL1Commitment: oldCommitment,
		newL1Commitment: newCommitment,
		resultCh:        resultChannel,
	}, resultChannel
}

func (msrT *ChainFetchStateDiff) GetOldStateIndex() uint32 {
	return msrT.oldStateIndex
}

func (msrT *ChainFetchStateDiff) GetNewStateIndex() uint32 {
	return msrT.newStateIndex
}

func (msrT *ChainFetchStateDiff) GetOldL1Commitment() *state.L1Commitment {
	return msrT.oldL1Commitment
}

func (msrT *ChainFetchStateDiff) GetNewL1Commitment() *state.L1Commitment {
	return msrT.newL1Commitment
}

func (msrT *ChainFetchStateDiff) IsValid() bool {
	return msrT.context.Err() == nil
}

func (msrT *ChainFetchStateDiff) Respond(theState *ChainFetchStateDiffResults) {
	if msrT.IsValid() && !msrT.IsResultChClosed() {
		msrT.resultCh <- theState
		msrT.closeResultCh()
	}
}

func (msrT *ChainFetchStateDiff) IsResultChClosed() bool {
	return msrT.resultCh == nil
}

func (msrT *ChainFetchStateDiff) closeResultCh() {
	close(msrT.resultCh)
	msrT.resultCh = nil
}
