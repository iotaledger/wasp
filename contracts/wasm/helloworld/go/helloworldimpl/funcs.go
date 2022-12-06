// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package helloworldimpl

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

func funcHelloWorld(ctx wasmlib.ScFuncContext, _ *HelloWorldContext) {
	ctx.Log("Hello, world!")
}

func viewGetHelloWorld(_ wasmlib.ScViewContext, f *GetHelloWorldContext) {
	f.Results.HelloWorld().SetValue("Hello, world!")
}
