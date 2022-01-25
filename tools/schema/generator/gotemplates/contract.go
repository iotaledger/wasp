package gotemplates

var contractGo = map[string]string{
	// *******************************
	"contract.go": `
$#emit goPackage
$#if funcs emitContract
`,
	// *******************************
	"emitContract": `

$#emit importWasmLib
$#each func FuncNameCall

type Funcs struct{}

var ScFuncs Funcs
$#each func FuncNameForCall
$#if core coreOnload
`,
	// *******************************
	"FuncNameCall": `
$#emit setupInitFunc

type $FuncName$+Call struct {
	Func    *wasmlib.Sc$initFunc$Kind
$#if param MutableFuncNameParams
$#if result ImmutableFuncNameResults
}
`,
	// *******************************
	"MutableFuncNameParams": `
	Params  Mutable$FuncName$+Params
`,
	// *******************************
	"ImmutableFuncNameResults": `
	Results Immutable$FuncName$+Results
`,
	// *******************************
	"FuncNameForCall": `
$#emit setupInitFunc

func (sc Funcs) $FuncName(ctx wasmlib.Sc$Kind$+CallContext) *$FuncName$+Call {
$#set thisView f.Func
$#if func setThisView
$#set complex $false
$#if param setComplex
$#if result setComplex
$#if complex initComplex initSimple
}
`,
	// *******************************
	"setThisView": `
$#set thisView &f.Func.ScView
`,
	// *******************************
	"setComplex": `
$#set complex $true
`,
	// *******************************
	"coreOnload": `

func OnLoad() {
	exports := wasmlib.NewScExports()
$#each func coreExportFunc
}
`,
	// *******************************
	"coreExportFunc": `
	exports.Add$Kind($Kind$FuncName, wasmlib.$Kind$+Error)
`,
	// *******************************
	"initComplex": `
	f := &$FuncName$+Call{Func: wasmlib.NewSc$initFunc$Kind(ctx, HScName, H$Kind$FuncName)}
$#if param initParams
$#if result initResults
	return f
`,
	// *******************************
	"initParams": `
	f.Params.proxy = wasmlib.NewCallParamsProxy($thisView)
`,
	// *******************************
	"initResults": `
	wasmlib.NewCallResultsProxy($thisView, &f.Results.proxy)
`,
	// *******************************
	"initSimple": `
	return &$FuncName$+Call{Func: wasmlib.NewSc$initFunc$Kind(ctx, HScName, H$Kind$FuncName)}
`,
}
