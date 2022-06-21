// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/execution"
)

type sandboxView struct {
	SandboxBase
}

func NewSandboxView(ctx execution.WaspContext) iscp.SandboxView {
	ret := &sandboxView{}
	ret.Ctx = ctx
	return ret
}

func (s *sandboxView) State() kv.KVStoreReader {
	return s.Ctx.StateReader()
}
