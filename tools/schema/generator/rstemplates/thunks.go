// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var thunksRs = map[string]string{
	// *******************************
	"thunks.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

use wasmlib::*;
use $package::*;
use crate::*;

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
$#each func libThunk
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
$#emit alignCalculate

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
    pub events:$align $Package$+Events,
`,
	// *******************************
	"PackageEventsInit": `
$#if events PackageEventsInitExist
`,
	// *******************************
	"PackageEventsInitExist": `
        events:$align $Package$+Events {},
`,
	// *******************************
	"ImmutableFuncNameParams": `
    pub params:$align Immutable$FuncName$+Params,
`,
	// *******************************
	"ImmutableFuncNameParamsInit": `
        params:$align Immutable$FuncName$+Params::new(),
`,
	// *******************************
	"MutableFuncNameResults": `
    pub results: Mutable$FuncName$+Results,
`,
	// *******************************
	"MutableFuncNameResultsInit": `
        results: Mutable$FuncName$+Results::new(),
`,
	// *******************************
	"PackageState": `
$#if func MutablePackageState
$#if view ImmutablePackageState
`,
	// *******************************
	"MutablePackageState": `
    pub state:$salign Mutable$Package$+State,
`,
	// *******************************
	"ImmutablePackageState": `
    pub state:$salign Immutable$Package$+State,
`,
	// *******************************
	"PackageStateInit": `
$#if func MutablePackageStateInit
$#if view ImmutablePackageStateInit
`,
	// *******************************
	"MutablePackageStateInit": `
        state:$salign Mutable$Package$+State::new(),
`,
	// *******************************
	"ImmutablePackageStateInit": `
        state:$salign Immutable$Package$+State::new(),
`,
	// *******************************
	"returnResultDict": `
    ctx.results(&f.results.proxy);
`,
	// *******************************
	"requireMandatory": `
    ctx.require(f.params.$fld_name().exists(), "missing mandatory param: $fldName");
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
