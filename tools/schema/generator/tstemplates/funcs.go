package tstemplates

var funcsTs = map[string]string{
	// *******************************
	"funcs.ts": `
use wasmlib::*;

use crate::*;
$#if structs useStructs
$#if typedefs useTypeDefs
$#each func funcSignature
`,
	// *******************************
	"funcSignature": `

pub fn $kind&+_$func_name(ctx: &Sc$Kind$+Context, f: &$FuncName$+Context) {
$#emit init$FuncName
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
	"initGetOwner": `
    f.results.owner().set_value(&f.state.owner().value());
`,
	// *******************************
	"initSetOwner": `
    f.state.owner().set_value(&f.params.owner().value());
`,
}
