// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var constsRs = map[string]string{
	// *******************************
	"consts.rs": `
#![allow(dead_code)]

$#if core useCrate useWasmLib

pub const SC_NAME        : &str = "$scName";
pub const SC_DESCRIPTION : &str = "$scDesc";
pub const HSC_NAME       : ScHname = ScHname(0x$hscName);
$#if params constParams
$#if results constResults
$#if state constState
$#if funcs constFuncs
`,
	// *******************************
	"constParams": `

$#set constPrefix PARAM_
$#each params constField
`,
	// *******************************
	"constResults": `

$#set constPrefix RESULT_
$#each results constField
`,
	// *******************************
	"constState": `

$#set constPrefix STATE_
$#each state constField
`,
	// *******************************
	"constFuncs": `

$#each func constFunc

$#each func constHFunc
`,
	// *******************************
	"constField": `
pub$crate const $constPrefix$FLD_NAME$fld_pad : &str = "$fldAlias";
`,
	// *******************************
	"constFunc": `
pub$crate const $KIND$+_$FUNC_NAME$func_pad : &str = "$funcName";
`,
	// *******************************
	"constHFunc": `
pub$crate const H$KIND$+_$FUNC_NAME$func_pad : ScHname = ScHname(0x$hFuncName);
`,
}
