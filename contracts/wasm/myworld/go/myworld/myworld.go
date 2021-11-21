// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package myworld

import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"

func funcDepositTreasure(ctx wasmlib.ScFuncContext, f *DepositTreasureContext) {
}

func funcInit(ctx wasmlib.ScFuncContext, f *InitContext) {
	if f.Params.Owner().Exists() {
		f.State.Owner().SetValue(f.Params.Owner().Value())
		return
	}
	f.State.Owner().SetValue(ctx.ContractCreator())
}

func funcSetOwner(ctx wasmlib.ScFuncContext, f *SetOwnerContext) {
	f.State.Owner().SetValue(f.Params.Owner().Value())
}

func viewGetAllTreasures(ctx wasmlib.ScViewContext, f *GetAllTreasuresContext) {
}

func viewGetOwner(ctx wasmlib.ScViewContext, f *GetOwnerContext) {
	f.Results.Owner().SetValue(f.State.Owner().Value())
}
