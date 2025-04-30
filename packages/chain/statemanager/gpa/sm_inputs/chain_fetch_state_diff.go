package sm_inputs

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
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

func NewChainFetchStateDiff(ctx context.Context, prevAnchor, nextAnchor *isc.StateAnchor) (*ChainFetchStateDiff, <-chan *ChainFetchStateDiffResults) {
	if prevAnchor == nil {
		// Only the current state is needed, if prevAO is unknown.
		prevAnchor = nextAnchor
	}
	oldCommitment, err := transaction.L1CommitmentFromAnchor(prevAnchor)
	if err != nil {
		panic(fmt.Errorf("cannot make L1 commitment from previous anchor, error: %w", err))
	}
	newCommitment, err := transaction.L1CommitmentFromAnchor(nextAnchor)
	if err != nil {
		panic(fmt.Errorf("cannot make L1 commitment from next anchor, error: %w", err))
	}
	resultChannel := make(chan *ChainFetchStateDiffResults, 1)
	return &ChainFetchStateDiff{
		context:         ctx,
		oldStateIndex:   prevAnchor.GetStateIndex(),
		newStateIndex:   nextAnchor.GetStateIndex(),
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
