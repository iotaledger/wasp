package gotemplates

var libGo = map[string]string{
	// *******************************
	"lib.go": `
//nolint:dupl
$#emit goHeader

func OnLoad() {
	exports := wasmlib.NewScExports()
$#each func libExportFunc

	for i, key := range keyMap {
		idxMap[i] = key.KeyID()
	}
}
$#each func libThunk
`,
	// *******************************
	"libExportFunc": `
	exports.Add$Kind($Kind$FuncName,$funcPad $kind$FuncName$+Thunk)
`,
	// *******************************
	"libThunk": `

type $FuncName$+Context struct {
$#if func PackageEvents
$#if param ImmutableFuncNameParams
$#if result MutableFuncNameResults
$#if func MutablePackageState
$#if view ImmutablePackageState
}

func $kind$FuncName$+Thunk(ctx wasmlib.Sc$Kind$+Context) {
	ctx.Log("$package.$kind$FuncName")
$#emit accessCheck
	f := &$FuncName$+Context{
$#if param ImmutableFuncNameParamsInit
$#if result MutableFuncNameResultsInit
$#if func MutablePackageStateInit
$#if view ImmutablePackageStateInit
	}
$#each mandatory requireMandatory
	$kind$FuncName(ctx, f)
	ctx.Log("$package.$kind$FuncName ok")
}
`,
	// *******************************
	"PackageEvents": `
$#if events PackageEventsExist
`,
	// *******************************
	"PackageEventsExist": `
	Events  $Package$+Events
`,
	// *******************************
	"ImmutableFuncNameParams": `
	Params  Immutable$FuncName$+Params
`,
	// *******************************
	"ImmutableFuncNameParamsInit": `
		Params: Immutable$FuncName$+Params{
			id: wasmlib.OBJ_ID_PARAMS,
		},
`,
	// *******************************
	"MutableFuncNameResults": `
	Results Mutable$FuncName$+Results
`,
	// *******************************
	"MutableFuncNameResultsInit": `
		Results: Mutable$FuncName$+Results{
			id: wasmlib.OBJ_ID_RESULTS,
		},
`,
	// *******************************
	"MutablePackageState": `
	State   Mutable$Package$+State
`,
	// *******************************
	"MutablePackageStateInit": `
		State: Mutable$Package$+State{
			id: wasmlib.OBJ_ID_STATE,
		},
`,
	// *******************************
	"ImmutablePackageState": `
	State   Immutable$Package$+State
`,
	// *******************************
	"ImmutablePackageStateInit": `
		State: Immutable$Package$+State{
			id: wasmlib.OBJ_ID_STATE,
		},
`,
	// *******************************
	"requireMandatory": `
	ctx.Requiref(f.Params.$FldName().Exists(), "missing mandatory $fldName")
`,

	// *******************************
	"accessCheck": `
$#set accessFinalize accessOther
$#emit caseAccess$funcAccess
$#emit $accessFinalize
`,
	// *******************************
	"caseAccess": `
$#set accessFinalize accessDone
`,
	// *******************************
	"caseAccessself": `
$#if funcAccessComment accessComment
	ctx.Requiref(ctx.Caller() == ctx.AccountID(), "no permission")

$#set accessFinalize accessDone
`,
	// *******************************
	"caseAccesschain": `
$#if funcAccessComment accessComment
	ctx.Requiref(ctx.Caller() == ctx.ChainOwnerID(), "no permission")

$#set accessFinalize accessDone
`,
	// *******************************
	"caseAccesscreator": `
$#if funcAccessComment accessComment
	ctx.Requiref(ctx.Caller() == ctx.ContractCreator(), "no permission")

$#set accessFinalize accessDone
`,
	// *******************************
	"accessOther": `
$#if funcAccessComment accessComment
	access := ctx.State().GetAgentID(wasmlib.Key("$funcAccess"))
	ctx.Requiref(access.Exists(), "access not set: $funcAccess")
	ctx.Requiref(ctx.Caller() == access.Value(), "no permission")

`,
	// *******************************
	"accessDone": `
`,
	// *******************************
	"accessComment": `

	$funcAccessComment
`,
}
