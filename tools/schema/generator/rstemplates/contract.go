package rstemplates

var contractRs = map[string]string{
	// *******************************
	"contract.rs": `
#![allow(dead_code)]

use std::ptr;

$#if core useCrate useWasmLib
$#if core useCoreContract contractUses
$#each func FuncNameCall

pub struct ScFuncs {
}

impl ScFuncs {
$#set separator $false
$#each func FuncNameForCall
}
`,
	// *******************************
	"contractUses": `

use crate::consts::*;
$#if params useParams
$#if results useResults
`,
	// *******************************
	"FuncNameCall": `
$#emit setupInitFunc

pub struct $FuncName$+Call {
	pub func: Sc$initFunc$Kind,
$#if param MutableFuncNameParams
$#if result ImmutableFuncNameResults
}
`,
	// *******************************
	"MutableFuncNameParams": `
	pub params: Mutable$FuncName$+Params,
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
    pub fn $func_name(_ctx: & dyn Sc$Kind$+CallContext) -> $FuncName$+Call {
$#set paramsID ptr::null_mut()
$#set resultsID ptr::null_mut()
$#if param setParamsID
$#if result setResultsID
$#if ptrs setPtrs noPtrs
    }
`,
	// *******************************
	"setPtrs": `
        let mut f = $FuncName$+Call {
            func: Sc$initFunc$Kind::new(HSC_NAME, H$KIND$+_$FUNC_NAME),
$#if param FuncNameParamsInit
$#if result FuncNameResultsInit
        };
        f.func.set_ptrs($paramsID, $resultsID);
        f
`,
	// *******************************
	"FuncNameParamsInit": `
            params: Mutable$FuncName$+Params { id: 0 },
`,
	// *******************************
	"FuncNameResultsInit": `
            results: Immutable$FuncName$+Results { id: 0 },
`,
	// *******************************
	"setParamsID": `
$#set paramsID &mut f.params.id
`,
	// *******************************
	"setResultsID": `
$#set resultsID &mut f.results.id
`,
	// *******************************
	"noPtrs": `
        $FuncName$+Call {
            func: Sc$initFunc$Kind::new(HSC_NAME, H$KIND$+_$FUNC_NAME),
        }
`,
}
