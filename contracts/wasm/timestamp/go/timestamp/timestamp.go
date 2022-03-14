// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package timestamp

import "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"

func funcNow(ctx wasmlib.ScFuncContext, f *NowContext) {
}

func viewGetTimestamp(ctx wasmlib.ScViewContext, f *GetTimestampContext) {
	f.Results.Timestamp().SetValue(ctx.Timestamp())
}
