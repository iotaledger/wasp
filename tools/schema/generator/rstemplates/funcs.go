// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var funcsRs = map[string]string{
	// *******************************
	"funcs.rs": `
use wasmlib::*;

use crate::*;
$#each func funcSignature
`,
	// *******************************
	"funcSignature": `

pub fn $kind$+_$func_name(ctx: &Sc$Kind$+Context, f: &$FuncName$+Context) {
$#emit init$Kind$FuncName
}
`,
	// *******************************
	"initFuncInit": `
    if f.params.owner().exists() {
        f.state.owner().set_value(&f.params.owner().value());
        return;
    }
    f.state.owner().set_value(&ctx.contract_creator());
`,
	// *******************************
	"initFuncSetOwner": `
    f.state.owner().set_value(&f.params.owner().value());
`,
	// *******************************
	"initViewGetOwner": `
    f.results.owner().set_value(&f.state.owner().value());
`,
}
