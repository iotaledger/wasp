package rstemplates

var funcsRs = map[string]string{
	// *******************************
	"funcs.rs": `
use wasmlib::*;

use crate::*;
use crate::contract::*;
$#if structs useStructs
$#if typedefs useTypeDefs
$#each func funcSignature
`,
	// *******************************
	"funcSignature": `

pub fn $kind&+_$func_name(ctx: &Sc$Kind$+Context, f: &$FuncName$+Context) {
$#func funcSignature
}
`,
	// *******************************
	"funcInit": `
    if f.params.owner().exists() {
        f.state.owner().set_value(&f.params.owner().value());
        return;
    }
    f.state.owner().set_value(&ctx.contract_creator());
`,
	// *******************************
	"getOwner": `
    f.results.owner().set_value(&f.state.owner().value());
`,
	// *******************************
	"setOwner": `
    f.state.owner().set_value(&f.params.owner().value());
`,
}
