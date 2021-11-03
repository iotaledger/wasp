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
$#each params constField
)
`,
	// *******************************
	"constResults": `

const (
$#each results constField
)
`,
	// *******************************
	"constState": `

const (
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
