package dist

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputChainDisabled struct {
	chainID isc.ChainID
}

func NewInputChainDisabled(chainID isc.ChainID) gpa.Input {
	return &inputChainDisabled{chainID: chainID}
}
