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

func (s sandboxView) MyBalances() coretypes.ColoredBalances {
	panic("implement me")
}

func (s sandboxView) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vmctx.CallView(contractHname, entryPoint, params)
}

func (s sandboxView) MyContractID() coretypes.ContractID {
	return s.vmctx.CurrentContractID()
}

func (s sandboxView) Event(msg string) {
	s.vmctx.Log().Infof("VMMSG contract %s '%s'", s.vmctx.CurrentContractID().String(), msg)
	s.vmctx.Publish(msg)
}

func (s sandboxView) Eventf(format string, args ...interface{}) {
	s.vmctx.Log().Infof("VMMSG: "+format, args...)
	s.vmctx.Publishf(format, args...)
}
