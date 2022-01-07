package tsclienttemplates

var funcsTs = map[string]string{
	// *******************************
	"funcs.ts": `
$#emit importEvents
$#emit importService
$#each events funcSignature
`,
	// *******************************
	"funcSignature": `

export function on$PkgName$EvtName(event: events.Event$EvtName): void {
}
`,
}
