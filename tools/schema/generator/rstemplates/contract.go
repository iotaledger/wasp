// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var contractRs = map[string]string{
	// *******************************
	"contract.rs": `
#![allow(dead_code)]
$#if core else useWasmLib
$#emit useCrate
$#if core useCoreContract
$#each func FuncNameCall

pub struct ScFuncs {
}

impl ScFuncs {
$#set separator $false
$#each func FuncNameForCall
}
`,
	// *******************************
	"FuncNameCall": `
$#emit alignCalculate
$#emit setupInitFunc

pub struct $FuncName$+Call<'a> {
    pub func:$falign Sc$initFunc$Kind<'a>,
$#if param MutableFuncNameParams
$#if result ImmutableFuncNameResults
}
`,
	// *******************************
	"MutableFuncNameParams": `
    pub params:$align Mutable$FuncName$+Params,
`,
	// *******************************
	"ImmutableFuncNameResults": `
    pub results: Immutable$FuncName$+Results,
`,
	// *******************************
	"FuncNameForCall": `
$#emit setupInitFunc
$#if separator newline
$#set separator $true
$#each funcComment _funcComment
    pub fn $func_name(ctx: &impl Sc$Kind$+CallContext) -> $FuncName$+Call {
$#if ptrs setPtrs noPtrs
    }
`,
	// *******************************
	"setPtrs": `
        let mut f = $FuncName$+Call {
            func:$falign Sc$initFunc$Kind::new(ctx, HSC_NAME, H$KIND$+_$FUNC_NAME),
$#if param FuncNameParamsInit
$#if result FuncNameResultsInit
        };
$#if param FuncNameParamsLink
$#if result FuncNameResultsLink
        f
`,
	// *******************************
	"FuncNameParamsInit": `
            params:$align Mutable$FuncName$+Params { proxy: Proxy::nil() },
`,
	// *******************************
	"FuncNameResultsInit": `
            results: Immutable$FuncName$+Results { proxy: Proxy::nil() },
`,
	// *******************************
	"FuncNameParamsLink": `
        Sc$initFunc$Kind::link_params(&mut f.params.proxy, &f.func);
`,
	// *******************************
	"FuncNameResultsLink": `
        Sc$initFunc$Kind::link_results(&mut f.results.proxy, &f.func);
`,
	// *******************************
	"noPtrs": `
        $FuncName$+Call {
            func: Sc$initFunc$Kind::new(ctx, HSC_NAME, H$KIND$+_$FUNC_NAME),
        }
`,
}
