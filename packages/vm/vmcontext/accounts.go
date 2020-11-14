package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (vmctx *VMContext) Transfer() coretypes.ColoredBalances {
	return vmctx.getCallContext().transfer
}
