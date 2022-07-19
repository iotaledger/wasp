// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package executiontime

import "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"

func funcF(ctx wasmlib.ScFuncContext, f *FContext) {
	n := f.Params.N().Value()
	x := uint32(0)
	y := uint32(0)

	for i := uint32(0); i < n; i++ {
		x++
		y = 3 * (x % 10)
	}

	f.Results.N().SetValue(y)
}
