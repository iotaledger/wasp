package smInputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type MempoolStateRequest struct {
	context         context.Context
	oldStateIndex   uint32
	newStateIndex   uint32
	oldL1Commitment *state.L1Commitment
	newL1Commitment *state.L1Commitment
	resultCh        chan<- *MempoolStateRequestResults
}

var _ gpa.Input = &MempoolStateRequest{}

func NewMempoolStateRequest(ctx context.Context, prevAO, nextAO *isc.AliasOutputWithID) (*MempoolStateRequest, <-chan *MempoolStateRequestResults) {
	oldCommitment, err := state.L1CommitmentFromAliasOutput(prevAO.GetAliasOutput())
	if err != nil {
		panic("Cannot make L1 commitment from previous alias output")
	}
	newCommitment, err := state.L1CommitmentFromAliasOutput(nextAO.GetAliasOutput())
	if err != nil {
		panic("Cannot make L1 commitment from next alias output")
	}
	resultChannel := make(chan *MempoolStateRequestResults, 1)
	return &MempoolStateRequest{
		context:         ctx,
		oldStateIndex:   prevAO.GetStateIndex(),
		newStateIndex:   nextAO.GetStateIndex(),
		oldL1Commitment: oldCommitment,
		newL1Commitment: newCommitment,
		resultCh:        resultChannel,
	}, resultChannel
}

func (msrT *MempoolStateRequest) GetOldStateIndex() uint32 {
	return msrT.oldStateIndex
}

func (msrT *MempoolStateRequest) GetNewStateIndex() uint32 {
	return msrT.newStateIndex
}

func (msrT *MempoolStateRequest) GetOldL1Commitment() *state.L1Commitment {
	return msrT.oldL1Commitment
}

func (msrT *MempoolStateRequest) GetNewL1Commitment() *state.L1Commitment {
	return msrT.newL1Commitment
}

func (msrT *MempoolStateRequest) IsValid() bool {
	return msrT.context.Err() == nil
}

func (msrT *MempoolStateRequest) Respond(theState *MempoolStateRequestResults) {
	if msrT.IsValid() && !msrT.IsResultChClosed() {
		msrT.resultCh <- theState
		msrT.closeResultCh()
	}
}

func (msrT *MempoolStateRequest) IsResultChClosed() bool {
	return msrT.resultCh == nil
}

func (msrT *MempoolStateRequest) closeResultCh() {
	close(msrT.resultCh)
	msrT.resultCh = nil
}
