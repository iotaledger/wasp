package inputs

import (
	"context"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type ConsensusBlockProduced struct {
	context    context.Context
	stateDraft state.StateDraft
	resultCh   chan<- state.Block
}

var _ gpa.Input = &ConsensusBlockProduced{}

func NewConsensusBlockProduced(ctx context.Context, stateDraft state.StateDraft) (*ConsensusBlockProduced, <-chan state.Block) {
	resultChannel := make(chan state.Block, 1)
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

func (cbpT *ConsensusBlockProduced) Respond(block state.Block) {
	if cbpT.IsValid() && !cbpT.isResultChClosed() {
		cbpT.resultCh <- block
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
