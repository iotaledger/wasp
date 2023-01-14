// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var thunksTs = map[string]string{
	// *******************************
	"thunks.ts": `
$#emit importWasmLib
import * as sc from '../$package/index';
import * as impl from './index'

const exportMap = new wasmlib.ScExportMap(
    [
$#each func libExportName
    ],
    [
$#each func libExportFunc
    ],
    [
$#each func libExportView
    ]);

export function onDispatch(index: i32): void {
    exportMap.dispatch(index);
}
$#each func libThunk
`,
	// *******************************
	"libExportName": `
        sc.$Kind$FuncName,
`,
	// *******************************
	"libExportFunc": `
$#if func libExportFuncThunk
`,
	// *******************************
	"libExportFuncThunk": `
        $kind$FuncName$+Thunk,
`,
	// *******************************
	"libExportView": `
$#if view libExportViewThunk
`,
	// *******************************
	"libExportViewThunk": `
        $kind$FuncName$+Thunk,
`,
	// *******************************
	"libThunk": `

function $kind$FuncName$+Thunk(ctx: wasmlib.Sc$Kind$+Context): void {
    ctx.log('$package.$kind$FuncName');
    let f = new sc.$FuncName$+Context();
$#if result initResultsDict
$#emit accessCheck
$#each mandatory requireMandatory
    impl.$kind$FuncName(ctx, f);
$#if result returnResultDict
    ctx.log('$package.$kind$FuncName ok');
}
`,
	// *******************************
	"initResultsDict": `
    const results = new wasmlib.ScDict(null);
    f.results = new sc.Mutable$FuncName$+Results(results.asProxy());
`,
	// *******************************
	"returnResultDict": `
    ctx.results(results);
`,
	// *******************************
	"requireMandatory": `
    ctx.require(f.params.$fldName().exists(), 'missing mandatory param: $fldName');
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
    ctx.require(ctx.caller().equals(ctx.accountID()), 'no permission');

$#set accessFinalize accessDone
`,
	// *******************************
	"caseAccesschain": `

$#each funcAccessComment _funcAccessComment
    ctx.require(ctx.caller().equals(ctx.chainOwnerID()), 'no permission');

$#set accessFinalize accessDone
`,
	// *******************************
	"accessOther": `

$#each funcAccessComment _funcAccessComment
    const access = f.state.$funcAccess();
    ctx.require(access.exists(), 'access not set: $funcAccess');
    ctx.require(ctx.caller().equals(access.value()), 'no permission');

`,
	// *******************************
	"accessDone": `
`,
}
