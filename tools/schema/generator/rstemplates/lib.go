// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var libRs = map[string]string{
	// *******************************
	"lib.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

use $package::*;
use wasmlib::*;
use wasmvmhost::*;

use crate::consts::*;
$#set moduleName events
$#if events useModule
$#set moduleName params
$#if params useModule
$#set moduleName results
$#if results useModule
$#set moduleName state
$#if state useModule
$#set moduleName structs
$#if structs useModule
$#set moduleName typedefs
$#if typedefs useModule

mod consts;
mod contract;
$#set moduleName events
$#if events modModule
$#set moduleName params
$#if params modModule
$#set moduleName results
$#if results modModule
$#set moduleName state
$#if state modModule
$#set moduleName structs
$#if structs modModule
$#set moduleName typedefs
$#if typedefs modModule

mod $package;

const EXPORT_MAP: ScExportMap = ScExportMap {
    names: &[
$#each func libExportName
    ],
    funcs: &[
$#each func libExportFunc
    ],
    views: &[
$#each func libExportView
    ],
};

pub fn on_dispatch(index: i32) {
    EXPORT_MAP.dispatch(index);
}

#[no_mangle]
fn on_call(index: i32) {
    WasmVmHost::connect();
    on_dispatch(index);
}

#[no_mangle]
fn on_load() {
    WasmVmHost::connect();
    on_dispatch(-1);
}
$#each func libThunk
`,
	// *******************************
	"useModule": `
use crate::$moduleName::*;
`,
	// *******************************
	"modModule": `
mod $moduleName;
`,
	// *******************************
	"libExportName": `
        $KIND$+_$FUNC_NAME,
`,
	// *******************************
	"libExportFunc": `
$#if func libExportFuncThunk
`,
	// *******************************
	"libExportFuncThunk": `
        $kind$+_$func_name$+_thunk,
`,
	// *******************************
	"libExportView": `
$#if view libExportViewThunk
`,
	// *******************************
	"libExportViewThunk": `
        $kind$+_$func_name$+_thunk,
`,
	// *******************************
	"libThunk": `

pub struct $FuncName$+Context {
$#if func PackageEvents
$#if param ImmutableFuncNameParams
$#if result MutableFuncNameResults
$#if state PackageState
}

fn $kind$+_$func_name$+_thunk(ctx: &Sc$Kind$+Context) {
    ctx.log("$package.$kind$FuncName");
    let f = $FuncName$+Context {
$#if func PackageEventsInit
$#if param ImmutableFuncNameParamsInit
$#if result MutableFuncNameResultsInit
$#if state PackageStateInit
    };
$#emit accessCheck
$#each mandatory requireMandatory
    $kind$+_$func_name(ctx, &f);
$#if result returnResultDict
    ctx.log("$package.$kind$FuncName ok");
}
`,
	// *******************************
	"PackageEvents": `
$#if events PackageEventsExist
`,
	// *******************************
	"PackageEventsExist": `
    events:  $Package$+Events,
`,
	// *******************************
	"PackageEventsInit": `
$#if events PackageEventsInitExist
`,
	// *******************************
	"PackageEventsInitExist": `
        events:  $Package$+Events {},
`,
	// *******************************
	"ImmutableFuncNameParams": `
    params: Immutable$FuncName$+Params,
`,
	// *******************************
	"ImmutableFuncNameParamsInit": `
        params: Immutable$FuncName$+Params { proxy: params_proxy() },
`,
	// *******************************
	"MutableFuncNameResults": `
    results: Mutable$FuncName$+Results,
`,
	// *******************************
	"MutableFuncNameResultsInit": `
        results: Mutable$FuncName$+Results { proxy: results_proxy() },
`,
	// *******************************
	"PackageState": `
$#if func MutablePackageState
$#if view ImmutablePackageState
`,
	// *******************************
	"MutablePackageState": `
    state: Mutable$Package$+State,
`,
	// *******************************
	"ImmutablePackageState": `
    state: Immutable$Package$+State,
`,
	// *******************************
	"PackageStateInit": `
$#if func MutablePackageStateInit
$#if view ImmutablePackageStateInit
`,
	// *******************************
	"MutablePackageStateInit": `
        state: Mutable$Package$+State { proxy: state_proxy() },
`,
	// *******************************
	"ImmutablePackageStateInit": `
        state: Immutable$Package$+State { proxy: state_proxy() },
`,
	// *******************************
	"returnResultDict": `
    ctx.results(&f.results.proxy.kv_store);
`,
	// *******************************
	"requireMandatory": `
    ctx.require(f.params.$fld_name().exists(), "missing mandatory $fldName");
`,
	// *******************************
	"accessCheck": `
$#set accessFinalize accessOther
$#emit caseAccess$funcAccess
$#emit $accessFinalize
`,
	// *******************************
	"caseAccess": `
$#set accessFinalize accessDone
`,
	// *******************************
	"caseAccessself": `

$#each funcAccessComment _funcAccessComment
    ctx.require(ctx.caller() == ctx.account_id(), "no permission");

$#set accessFinalize accessDone
`,
	// *******************************
	"caseAccesschain": `

$#each funcAccessComment _funcAccessComment
    ctx.require(ctx.caller() == ctx.chain_owner_id(), "no permission");

$#set accessFinalize accessDone
`,
	// *******************************
	"accessOther": `

$#each funcAccessComment _funcAccessComment
    let access = f.state.$func_access();
    ctx.require(access.exists(), "access not set: $funcAccess");
    ctx.require(ctx.caller() == access.value(), "no permission");

`,
	// *******************************
	"accessDone": `
`,
}
