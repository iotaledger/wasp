package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

var (
	NewSandbox     func(vmctx *VMContext) iscp.Sandbox
	NewSandboxView func(vmctx *VMContext) iscp.SandboxView
)
