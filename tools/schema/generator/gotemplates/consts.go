package gotemplates

var constsGo = map[string]string{
	// *******************************
	"consts.go": `
$#emit goHeader

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
$#set constPrefix Param
$#each params constField
)
`,
	// *******************************
	"constResults": `

const (
$#set constPrefix Result
$#each results constField
)
`,
	// *******************************
	"constState": `

const (
$#set constPrefix State
$#each state constField
)
`,
	// *******************************
	"constField": `
	$constPrefix$FldName$fldPad = "$fldAlias"
`,
	// *******************************
	"constFunc": `
	$Kind$FuncName$funcPad = "$funcName"
`,
	// *******************************
	"constHFunc": `
	H$Kind$FuncName$funcPad = wasmlib.ScHname(0x$funcHname)
`,
}
