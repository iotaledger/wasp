package tsclienttemplates

var funcsTs = map[string]string{
	// *******************************
	"funcs.ts": `
import * as events from "./events"
import * as service from "./service"
$#each events funcSignature
`,
	// *******************************
	"funcSignature": `

export function on$PkgName$EvtName(event: events.Event$EvtName): void {
}
`,
}
