// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package sandbox

import (
	"github.com/iotaledger/wasp/packages/coret"
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

func (s sandboxView) MyBalances() coret.ColoredBalances {
	return s.vmctx.GetMyBalances()
}

func (s sandboxView) Call(contractHname coret.Hname, entryPoint coret.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vmctx.Call(contractHname, entryPoint, params, nil)
}

func (s sandboxView) MyContractID() coret.ContractID {
	return s.vmctx.CurrentContractID()
}

func (s sandboxView) Event(msg string) {
	s.vmctx.EventPublisher().Publish(msg)
}

func (s sandboxView) Eventf(format string, args ...interface{}) {
	s.vmctx.EventPublisher().Publishf(format, args...)
}
