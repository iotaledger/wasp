package smInputs

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type ChainReceiveConfirmedAliasOutput struct {
	stateOutput *isc.AliasOutputWithID
}

var _ gpa.Input = &ChainReceiveConfirmedAliasOutput{}

func NewChainReceiveConfirmedAliasOutput(output *isc.AliasOutputWithID) *ChainReceiveConfirmedAliasOutput {
	return &ChainReceiveConfirmedAliasOutput{stateOutput: output}
}

func (crcaoT *ChainReceiveConfirmedAliasOutput) GetStateOutput() *isc.AliasOutputWithID {
	return crcaoT.stateOutput
}
