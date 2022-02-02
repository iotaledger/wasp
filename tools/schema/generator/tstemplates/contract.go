package tstemplates

var contractTs = map[string]string{
	// *******************************
	"contract.ts": `
$#emit tsImports
$#each func FuncNameCall

export class ScFuncs {
$#set separator $false
$#each func FuncNameForCall
}
`,
	// *******************************
	"FuncNameCall": `
$#emit setupInitFunc

export class $FuncName$+Call {
	func: wasmlib.Sc$initFunc$Kind = new wasmlib.Sc$initFunc$Kind(sc.HScName, sc.H$Kind$FuncName);
$#if param MutableFuncNameParams
$#if result ImmutableFuncNameResults
}
$#if core else FuncNameContext
`,
	// *******************************
	"FuncNameContext": `

export class $FuncName$+Context {
$#if func PackageEvents
$#if param ImmutableFuncNameParams
$#if result MutableFuncNameResults
$#if func MutablePackageState
$#if view ImmutablePackageState
}
`,
	// *******************************
	"PackageEvents": `
$#if events PackageEventsExist
`,
	// *******************************
	"PackageEventsExist": `
	events: sc.$Package$+Events = new sc.$Package$+Events();
`,
	// *******************************
	"ImmutableFuncNameParams": `
	params: sc.Immutable$FuncName$+Params = new sc.Immutable$FuncName$+Params();
`,
	// *******************************
	"MutableFuncNameParams": `
	params: sc.Mutable$FuncName$+Params = new sc.Mutable$FuncName$+Params();
`,
	// *******************************
	"ImmutableFuncNameResults": `
	results: sc.Immutable$FuncName$+Results = new sc.Immutable$FuncName$+Results();
`,
	// *******************************
	"MutableFuncNameResults": `
	results: sc.Mutable$FuncName$+Results = new sc.Mutable$FuncName$+Results();
`,
	// *******************************
	"ImmutablePackageState": `
	state: sc.Immutable$Package$+State = new sc.Immutable$Package$+State();
`,
	// *******************************
	"MutablePackageState": `
	state: sc.Mutable$Package$+State = new sc.Mutable$Package$+State();
`,
	// *******************************
	"FuncNameForCall": `
$#emit setupInitFunc
$#if separator newline
$#set separator $true
    static $funcName(ctx: wasmlib.Sc$Kind$+CallContext): $FuncName$+Call {
$#set paramsID null
$#set resultsID null
$#if param setParamsID
$#if result setResultsID
$#if ptrs setPtrs noPtrs
    }
`,
	// *******************************
	"setPtrs": `
        let f = new $FuncName$+Call();
        f.func.setPtrs($paramsID, $resultsID);
        return f;
`,
	// *******************************
	"setParamsID": `
$#set paramsID f.params
`,
	// *******************************
	"setResultsID": `
$#set resultsID f.results
`,
	// *******************************
	"noPtrs": `
        return new $FuncName$+Call();
`,
}
