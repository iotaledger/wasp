package gotemplates

var funcsGo = map[string]string{
	// *******************************
	"funcs.go": `
$#emit goHeader
$#each func funcSignature
`,
	// *******************************
	"funcSignature": `

func $kind$FuncName(ctx wasmlib.Sc$Kind$+Context, f *$FuncName$+Context) {
$#emit init$FuncName
}
`,
	// *******************************
	"initFuncInit": `
    if f.Params.Owner().Exists() {
        f.State.Owner().SetValue(f.Params.Owner().Value())
        return
    }
    f.State.Owner().SetValue(ctx.ContractCreator())
`,
	// *******************************
	"intGetOwner": `
	f.Results.Owner().SetValue(f.State.Owner().Value())
`,
	// *******************************
	"initSetOwner": `
	f.State.Owner().SetValue(f.Params.Owner().Value())
`,
}
