// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var funcsGo = map[string]string{
	// *******************************
	"funcs.go": `
package $package$+impl

import (
	"$module/go/$package$+impl"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)
$#each func funcSignature
`,
	// *******************************
	"funcSignature": `

func $kind$FuncName(ctx wasmlib.Sc$Kind$+Context, f *$FuncName$+Context) {
$#emit init$Kind$FuncName
}
`,
	// *******************************
	"initFuncInit": `
	if f.Params.Owner().Exists() {
		f.State.Owner().SetValue(f.Params.Owner().Value())
		return
	}
	f.State.Owner().SetValue(ctx.RequestSender())
`,
	// *******************************
	"initFuncSetOwner": `
	f.State.Owner().SetValue(f.Params.Owner().Value())
`,
	// *******************************
	"initViewGetOwner": `
	f.Results.Owner().SetValue(f.State.Owner().Value())
`,
}
