package tsclienttemplates

var funcsTs = map[string]string{
	// *******************************
	"funcs.ts": `
import * as events from "./events"
import * as service from "./service"

const client = new BasicClient(config);
const $pkgName$+Service = new service.$PkgName$+Service(client, config.ChainId);
$#each events funcSignature
`,
	// *******************************
	"funcSignature": `

export function on$PkgName$EvtName(event: events.Event$EvtName): void {
}
`,
}
