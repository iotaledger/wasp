package gotemplates

var contractGo = map[string]string{
	// *******************************
	"contract.go": `
$#emit goHeader
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
$#set paramsID nil
$#set resultsID nil
$#if param setParamsID
$#if result setResultsID
$#if ptrs setPtrs noPtrs
}
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
	exports.Add$Kind($Kind$FuncName, wasmlib.$Kind$+ErrorStr)
`,
	// *******************************
	"setPtrs": `
	f := &$FuncName$+Call{Func: wasmlib.NewSc$initFunc$Kind(ctx, HScName, H$Kind$FuncName$initMap)}
	f.Func.SetPtrs($paramsID, $resultsID)
	return f
`,
	// *******************************
	"setParamsID": `
$#set paramsID &f.Params.id
`,
	// *******************************
	"setResultsID": `
$#set resultsID &f.Results.id
`,
	// *******************************
	"noPtrs": `
	return &$FuncName$+Call{Func: wasmlib.NewSc$initFunc$Kind(ctx, HScName, H$Kind$FuncName$initMap)}
`,
}
