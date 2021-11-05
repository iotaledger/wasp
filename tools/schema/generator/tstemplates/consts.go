package tstemplates

var constsTs = map[string]string{
	// *******************************
	"consts.ts": `
$#emit tsHeader

const (
	ScName        = "$scName"
	ScDescription = "$scDesc"
	HScName       = wasmlib.ScHname(0x$hscName)
)
$#if params constParams
$#if results constResults
$#if state constState

const (
$#each func constFunc
)

const (
$#each func constHFunc
)
`,
	// *******************************
	"constParams": `

const (
$#set constPrefix "Param"
$#each params constField
)
`,
	// *******************************
	"constResults": `

const (
$#set constPrefix "Result"
$#each results constField
)
`,
	// *******************************
	"constState": `

const (
$#set constPrefix "State"
$#each state constField
)
`,
	// *******************************
	"constField": `
	$constPrefix$FldName = wasmlib.Key("$fldAlias")
`,
	// *******************************
	"constFunc": `
	$Kind$FuncName = "$funcName"
`,
	// *******************************
	"constHFunc": `
	H$Kind$FuncName = wasmlib.ScHname(0x$funcHName)
`,
}
