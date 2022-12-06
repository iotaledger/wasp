// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package timestampimpl

import "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"

func funcNow(_ wasmlib.ScFuncContext, _ *NowContext) {
}

func viewGetTimestamp(ctx wasmlib.ScViewContext, f *GetTimestampContext) {
	f.Results.Timestamp().SetValue(ctx.Timestamp())
}
