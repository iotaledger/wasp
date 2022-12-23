// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var contractTs = map[string]string{
	// *******************************
	"contract.ts": `
$#emit importWasmLib
$#emit importSc
$#each func FuncNameCall

export class ScFuncs {
$#set separator $false
$#each func FuncNameForCall
}
`,
	// *******************************
	"FuncNameCall": `
$#emit alignCalculate
$#emit setupInitFunc

export class $FuncName$+Call {
    func:$falign wasmlib.Sc$initFunc$Kind;
$#if param MutableFuncNameParams
$#if result ImmutableFuncNameResults

    public constructor(ctx: wasmlib.Sc$Kind$+CallContext) {
        this.func = new wasmlib.Sc$initFunc$Kind(ctx, sc.HScName, sc.H$Kind$FuncName);
    }
}
$#if core else FuncNameContext
`,
	// *******************************
	"FuncNameContext": `

export class $FuncName$+Context {
$#if func PackageEvents
$#if param ImmutableFuncNameParams
$#if result MutableFuncNameResults
$#if state PackageState
}
`,
	// *******************************
	"PackageEvents": `
$#if events PackageEventsExist
`,
	// *******************************
	"PackageEventsExist": `
    events:$align sc.$Package$+Events = new sc.$Package$+Events();
`,
	// *******************************
	"ImmutableFuncNameParams": `
    params:$align sc.Immutable$FuncName$+Params = new sc.Immutable$FuncName$+Params(wasmlib.paramsProxy());
`,
	// *******************************
	"MutableFuncNameParams": `
    params:$align sc.Mutable$FuncName$+Params = new sc.Mutable$FuncName$+Params(wasmlib.ScView.nilProxy);
`,
	// *******************************
	"ImmutableFuncNameResults": `
    results: sc.Immutable$FuncName$+Results = new sc.Immutable$FuncName$+Results(wasmlib.ScView.nilProxy);
`,
	// *******************************
	"MutableFuncNameResults": `
    results: sc.Mutable$FuncName$+Results = new sc.Mutable$FuncName$+Results(wasmlib.ScView.nilProxy);
`,
	// *******************************
	"PackageState": `
$#if func MutablePackageState
$#if view ImmutablePackageState
`,
	// *******************************
	"ImmutablePackageState": `
    state:$salign sc.Immutable$Package$+State = new sc.Immutable$Package$+State(wasmlib.ScState.proxy());
`,
	// *******************************
	"MutablePackageState": `
    state:$salign sc.Mutable$Package$+State = new sc.Mutable$Package$+State(wasmlib.ScState.proxy());
`,
	// *******************************
	"FuncNameForCall": `
$#emit setupInitFunc
$#if separator newline
$#set separator $true
$#each funcComment _funcComment
    static $funcName(ctx: wasmlib.Sc$Kind$+CallContext): $FuncName$+Call {
$#if ptrs setPtrs noPtrs
    }
`,
	// *******************************
	"setPtrs": `
        const f = new $FuncName$+Call(ctx);
$#if param initParams
$#if result initResults
        return f;
`,
	// *******************************
	"initParams": `
        f.params = new sc.Mutable$FuncName$+Params(wasmlib.newCallParamsProxy(f.func));
`,
	// *******************************
	"initResults": `
        f.results = new sc.Immutable$FuncName$+Results(wasmlib.newCallResultsProxy(f.func));
`,
	// *******************************
	"noPtrs": `
        return new $FuncName$+Call(ctx);
`,
}
