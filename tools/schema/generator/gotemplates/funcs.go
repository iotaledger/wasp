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
$#func funcSignature
}
`,
	// *******************************
	"funcInit": `
    if f.Params.Owner().Exists() {
        f.State.Owner().SetValue(f.Params.Owner().Value())
        return
    }
    f.State.Owner().SetValue(ctx.ContractCreator())
`,
	// *******************************
	"getOwner": `
	f.Results.Owner().SetValue(f.State.Owner().Value())
`,
	// *******************************
	"setOwner": `
	f.State.Owner().SetValue(f.Params.Owner().Value())
`,
}
