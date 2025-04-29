// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package sandbox implements the vm sandbox
package sandbox

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/execution"
)

type sandboxView struct {
	SandboxBase
}

func NewSandboxView(ctx execution.WaspCallContext) isc.SandboxView {
	return &sandboxView{
		SandboxBase: SandboxBase{
			Ctx: ctx,
		},
	}
}
