package vmcontext

import (
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

var NewSandbox func(vmctx *VMContext) vmtypes.Sandbox
