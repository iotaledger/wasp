package tstemplates

var funcsTs = map[string]string{
	// *******************************
	"funcs.ts": `
$#emit importWasmLib
$#emit importSc
$#each func funcSignature
`,
	// *******************************
	"funcSignature": `

export function $kind$+$FuncName(ctx: wasmlib.Sc$Kind$+Context, f: sc.$FuncName$+Context): void {
$#emit init$FuncName
}
`,
	// *******************************
	"initFuncInit": `
    if f.params.owner().exists() {
        f.state.owner().setValue(&f.params.owner().value());
        return;
    }
    f.state.owner().setValue(ctx.contractCreator());
`,
	// *******************************
	"initGetOwner": `
    f.results.owner().setValue(f.state.owner().value());
`,
	// *******************************
	"initSetOwner": `
    f.state.owner().setValue(f.params.owner().value());
`,
}
