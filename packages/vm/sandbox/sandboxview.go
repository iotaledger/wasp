package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func init() {
	vmcontext.NewSandboxView = newView
}

type sandboxView struct {
	vmctx *vmcontext.VMContext
}

func newView(vmctx *vmcontext.VMContext) vmtypes.SandboxView {
	return sandboxView{vmctx}
}

func (s sandboxView) Params() codec.ImmutableCodec {
	return s.vmctx.Params()
}

func (s sandboxView) State() codec.ImmutableMustCodec {
	return codec.NewMustCodec(s.vmctx)
}

func (s sandboxView) Accounts() coretypes.ColoredAccountsImmutable {
	return s.vmctx.Accounts()
}

func (s sandboxView) Call(contractIndex uint16, entryPoint coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vmctx.CallView(contractIndex, entryPoint, params)
}
