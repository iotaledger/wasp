package rstemplates

var contractRs = map[string]string{
	// *******************************
	"contract.rs": `
// @formatter:off

#![allow(dead_code)]

use std::ptr;

use wasmlib::*;
$#if core else ContractUses
$#each func FuncNameCall

pub struct ScFuncs {
}

impl ScFuncs {
$#each func FuncNameForCall
}

// @formatter:on
`,
	// *******************************
	"ContractUses": `

use crate::consts::*;
$#if params useParams
$#if results useResults
`,
	// *******************************
	"FuncNameCall": `

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
    pub fn $func_name(_ctx: & dyn Sc$Kind$+CallContext) -> $FuncName$+Call {
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
	"noPtrs": `
        $FuncName$+Call {
            func: Sc$initFunc$Kind::new(HSC_NAME, H$KIND$+_$FUNC_NAME),
        }
`,
}
