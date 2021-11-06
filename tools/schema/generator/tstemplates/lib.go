package tstemplates

var libTs = map[string]string{
	// *******************************
	"lib.ts": `
$#emit importWasmLib
$#emit importSc

export function on_call(index: i32): void {
    return wasmlib.onCall(index);
}

export function on_load(): void {
    let exports = new wasmlib.ScExports();
$#each func libExportFunc

    for (let i = 0; i < sc.keyMap.length; i++) {
        sc.idxMap[i] = wasmlib.Key32.fromString(sc.keyMap[i]);
    }
}
$#each func libThunk
`,
	// *******************************
	"libExportFunc": `
    exports.add$Kind(sc.$Kind$FuncName, $kind$FuncName$+Thunk);
`,
	// *******************************
	"libThunk": `

function $kind$FuncName$+Thunk(ctx: wasmlib.Sc$Kind$+Context): void {
	ctx.log("$package.$kind$FuncName");
$#func accessCheck
	let f = new sc.$FuncName$+Context();
$#if param ImmutableFuncNameParamsInit
$#if result MutableFuncNameResultsInit
    f.state.mapID = wasmlib.OBJ_ID_STATE;
$#each mandatory requireMandatory
	sc.$kind$FuncName(ctx, f);
	ctx.log("$package.$kind$FuncName ok");
}
`,
	// *******************************
	"ImmutableFuncNameParamsInit": `
    f.params.mapID = wasmlib.OBJ_ID_PARAMS;
`,
	// *******************************
	"MutableFuncNameResultsInit": `
    f.results.mapID = wasmlib.OBJ_ID_RESULTS;
`,
	// *******************************
	"requireMandatory": `
	ctx.require(f.params.$fldName().exists(), "missing mandatory $fldName");
`,
	// *******************************
	"grantForKey": `
	let access = ctx.state().getAgentID(wasmlib.Key32.fromString("$grant"));
	ctx.require(access.exists(), "access not set: $grant");
`,
	// *******************************
	"grantRequire": `
	ctx.require(ctx.caller().equals($grant), "no permission");

`,
}
