// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type sandboxView struct {
	Sandboxbase
}

func NewSandboxView(ctx execution.WaspContext) iscp.SandboxView {
	ret := &sandboxView{}
	ret.Ctx = ctx
	return ret
}

func (s *Sandboxbase) State() kv.KVStoreReader {
	return s.Ctx.StateReader()
}

func (s *sandboxView) Call(contractHname, entryPoint iscp.Hname, params dict.Dict) dict.Dict {
	s.Ctx.GasBurn(gas.BurnCodeCallContract)
	if params == nil {
		params = make(dict.Dict)
	}
	return s.Ctx.Call(contractHname, entryPoint, params, nil)
}
