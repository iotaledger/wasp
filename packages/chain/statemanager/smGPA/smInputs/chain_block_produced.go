package smInputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type ChainBlockProduced struct {
	context  context.Context
	block    state.Block
	resultCh chan<- error
}

var _ gpa.Input = &ChainBlockProduced{}

func NewChainBlockProduced(ctx context.Context, block state.Block) (*ChainBlockProduced, <-chan error) {
	resultChannel := make(chan error, 1)
	return &ChainBlockProduced{
		context:  ctx,
		block:    block,
		resultCh: resultChannel,
	}, resultChannel
}

func (cbpT *ChainBlockProduced) GetBlock() state.Block {
	return cbpT.block
}

func (cbpT *ChainBlockProduced) IsValid() bool {
	return cbpT.context.Err() == nil
}

func (cbpT *ChainBlockProduced) Respond(err error) {
	if cbpT.IsValid() && !cbpT.isResultChClosed() {
		cbpT.resultCh <- err
		cbpT.closeResultCh()
	}
}

func (cbpT *ChainBlockProduced) isResultChClosed() bool {
	return cbpT.resultCh == nil
}

func (cbpT *ChainBlockProduced) closeResultCh() {
	close(cbpT.resultCh)
	cbpT.resultCh = nil
}
