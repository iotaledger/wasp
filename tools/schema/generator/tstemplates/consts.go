package tstemplates

var constsTs = map[string]string{
	// *******************************
	"consts.ts": `
$#emit importWasmLib

export const ScName        = "$scName";
export const ScDescription = "$scDesc";
export const HScName       = new wasmlib.ScHname(0x$hscName);
$#if params constParams
$#if results constResults
$#if state constState

$#each func constFunc

$#each func constHFunc
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
	"constField": `
export const $constPrefix$FldName$fldPad = "$fldAlias";
`,
	// *******************************
	"constFunc": `
export const $Kind$FuncName$funcPad = "$funcName";
`,
	// *******************************
	"constHFunc": `
export const H$Kind$FuncName$funcPad = new wasmlib.ScHname(0x$funcHname);
`,
}
