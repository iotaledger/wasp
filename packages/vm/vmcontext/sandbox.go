package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

var (
	NewSandbox     func(vmctx *VMContext) coretypes.Sandbox
	NewSandboxView func(vmctx *VMContext) coretypes.SandboxView
)
