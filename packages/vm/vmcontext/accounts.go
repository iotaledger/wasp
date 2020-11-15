package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (vmctx *VMContext) Incoming() coretypes.ColoredBalances {
	return vmctx.getCallContext().transfer
}
