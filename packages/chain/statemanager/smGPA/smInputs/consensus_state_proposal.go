package smInputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type ConsensusStateProposal struct {
	context     context.Context
	aliasOutput *isc.AliasOutputWithID
	resultCh    chan<- interface{}
}

var _ gpa.Input = &ConsensusStateProposal{}

func NewConsensusStateProposal(ctx context.Context, aliasOutput *isc.AliasOutputWithID) (*ConsensusStateProposal, <-chan interface{}) {
	resultChannel := make(chan interface{}, 1)
	return &ConsensusStateProposal{
		context:     ctx,
		aliasOutput: aliasOutput,
		resultCh:    resultChannel,
	}, resultChannel
}

func (cspT *ConsensusStateProposal) GetAliasOutputWithID() *isc.AliasOutputWithID {
	return cspT.aliasOutput
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
