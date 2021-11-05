package rstemplates

var libRs = map[string]string{
	// *******************************
	"lib.rs": `
// @formatter:off

#![allow(dead_code)]
#![allow(unused_imports)]

use $package::*;
use wasmlib::*;
use wasmlib::host::*;

use crate::consts::*;
use crate::keys::*;
$#if params useParams
$#if results useResults
use crate::state::*;

mod consts;
mod contract;
mod keys;
$#if params modParams
$#if results modResults
mod state;
$#if structs modStructs
$#if typedefs modTypeDefs
mod $package;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
$#each func libExportFunc

    unsafe {
        for i in 0..KEY_MAP_LEN {
            IDX_MAP[i] = get_key_id_from_string(KEY_MAP[i]);
        }
    }
}
$#each func libThunk

// @formatter:on
`,
	// *******************************
	"libExportFunc": `
    exports.add_$kind($KIND$+_$FUNC_NAME, $kind$+_$func_name$+_thunk);
`,
	// *******************************
	"libThunk": `

pub struct $FuncName$+Context {
$#if param ImmutableFuncNameParams
$#if result MutableFuncNameResults
$#if func MutablePackageState
$#if view ImmutablePackageState
}

fn $kind$+_$func_name$+_thunk(ctx: &Sc$Kind$+Context) {
	ctx.log("$package.$kind$FuncName");
$#func accessCheck
	let f = $FuncName$+Context {
$#if param ImmutableFuncNameParamsInit
$#if result MutableFuncNameResultsInit
$#if func MutablePackageStateInit
$#if view ImmutablePackageStateInit
	};
$#each mandatory requireMandatory
	$kind$+_$func_name(ctx, &f);
	ctx.log("$package.$kind$FuncName ok");
}
`,
	// *******************************
	"ImmutableFuncNameParams": `
	params: Immutable$FuncName$+Params,
`,
	// *******************************
	"ImmutableFuncNameParamsInit": `
		params: Immutable$FuncName$+Params {
			id: OBJ_ID_PARAMS,
		},
`,
	// *******************************
	"MutableFuncNameResults": `
	results: Mutable$FuncName$+Results,
`,
	// *******************************
	"MutableFuncNameResultsInit": `
		results: Mutable$FuncName$+Results {
			id: OBJ_ID_RESULTS,
		},
`,
	// *******************************
	"MutablePackageState": `
	state: Mutable$Package$+State,
`,
	// *******************************
	"MutablePackageStateInit": `
		state: Mutable$Package$+State {
			id: OBJ_ID_STATE,
		},
`,
	// *******************************
	"ImmutablePackageState": `
	state: Immutable$Package$+State,
`,
	// *******************************
	"ImmutablePackageStateInit": `
		state: Immutable$Package$+State {
			id: OBJ_ID_STATE,
		},
`,
	// *******************************
	"requireMandatory": `
	ctx.require(f.params.$fld_name().exists(), "missing mandatory $fldName");
`,
	// *******************************
	"grantForKey": `
	let access = ctx.state().get_agent_id("$grant");
	ctx.require(access.exists(), "access not set: $grant");
`,
	// *******************************
	"grantRequire": `
	ctx.require(ctx.caller() == $grant, "no permission");

`,
}
