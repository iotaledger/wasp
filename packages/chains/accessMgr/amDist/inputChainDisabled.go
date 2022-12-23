package amDist

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputChainDisabled struct {
	chainID isc.ChainID
}

func NewInputChainDisabled(chainID isc.ChainID) gpa.Input {
	return &inputChainDisabled{chainID: chainID}
}
