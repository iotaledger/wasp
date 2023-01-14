// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var constsTs = map[string]string{
	// *******************************
	"consts.ts": `
$#emit importWasmTypes

export const ScName        = '$scName';
export const ScDescription = '$scDesc';
export const HScName       = new wasmtypes.ScHname(0x$hscName);
$#if params constParams
$#if results constResults
$#if state constState
$#if funcs constFuncs
`,
	// *******************************
	"constParams": `

$#set constPrefix Param
$#each params constField
`,
	// *******************************
	"constResults": `

$#set constPrefix Result
$#each results constField
`,
	// *******************************
	"constState": `

$#set constPrefix State
$#each state constField
`,
	// *******************************
	"constFuncs": `

$#each func constFunc

$#each func constHFunc
`,
	// *******************************
	"constField": `
export const $constPrefix$FldName$fldPad = '$fldAlias';
`,
	// *******************************
	"constFunc": `
export const $Kind$FuncName$funcPad = '$funcName';
`,
	// *******************************
	"constHFunc": `
export const H$Kind$FuncName$funcPad = new wasmtypes.ScHname(0x$hFuncName);
`,
}
