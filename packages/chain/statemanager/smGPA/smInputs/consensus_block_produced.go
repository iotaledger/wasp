package smInputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type ConsensusBlockProduced struct {
	context    context.Context
	stateDraft state.StateDraft
	resultCh   chan<- error
}

var _ gpa.Input = &ConsensusBlockProduced{}

func NewConsensusBlockProduced(ctx context.Context, stateDraft state.StateDraft) (*ConsensusBlockProduced, <-chan error) {
	resultChannel := make(chan error, 1)
	return &ConsensusBlockProduced{
		context:    ctx,
		stateDraft: stateDraft,
		resultCh:   resultChannel,
	}, resultChannel
}

func (cbpT *ConsensusBlockProduced) GetStateDraft() state.StateDraft {
	return cbpT.stateDraft
}

func (cbpT *ConsensusBlockProduced) IsValid() bool {
	return cbpT.context.Err() == nil
}

func (cbpT *ConsensusBlockProduced) Respond(err error) {
	if cbpT.IsValid() && !cbpT.isResultChClosed() {
		cbpT.resultCh <- err
		cbpT.closeResultCh()
	}
}

func (cbpT *ConsensusBlockProduced) isResultChClosed() bool {
	return cbpT.resultCh == nil
}

func (cbpT *ConsensusBlockProduced) closeResultCh() {
	close(cbpT.resultCh)
	cbpT.resultCh = nil
}
