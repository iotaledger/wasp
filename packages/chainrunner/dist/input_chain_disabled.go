package dist

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type inputChainDisabled struct{}

func NewInputChainDisabled() gpa.Input {
	return &inputChainDisabled{}
}
