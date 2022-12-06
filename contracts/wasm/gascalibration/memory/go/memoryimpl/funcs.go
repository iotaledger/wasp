// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package memoryimpl

import "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"

func funcF(ctx wasmlib.ScFuncContext, f *FContext) {
	n := f.Params.N().Value()
	store := make([]uint32, n)
	for i := uint32(0); i < n; i++ {
		store[i] = i
	}
}
