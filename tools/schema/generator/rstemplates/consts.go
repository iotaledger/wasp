package rstemplates

var constsRs = map[string]string{
	// *******************************
	"consts.rs": `
// @formatter:off

#![allow(dead_code)]

$#emit rsHeader

pub const SC_NAME:        &str = "$scName";
pub const SC_DESCRIPTION: &str = "$scDesc";
pub const HSC_NAME:       ScHname = ScHname(0x$hscName);
$#if params constParams
$#if results constResults
$#if state constState

$#each func constFunc

$#each func constHFunc

// @formatter:on
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
	"constField": `
pub const $constPrefix$FLD_NAME: &str = "$fldAlias";
`,
	// *******************************
	"constFunc": `
pub const $KIND$+_$FUNC_NAME:  &str = "$funcName";
`,
	// *******************************
	"constHFunc": `
pub const H$KIND$+_$FUNC_NAME: ScHname = ScHname(0x$funcHName);
`,
}
