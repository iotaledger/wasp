package clienttemplates

var appTs = map[string]string{
	// *******************************
	"app.ts": `
import * as wasmlib from "./wasmlib"
import * as events from "./events"
import * as service from "./service"

let $pkgName$+Service: service.$PkgName$+Service;

export function subscribeTo$Package$+Events(): void {
$#each events appOnEvent
}
`,
	// *******************************
	"appOnEvent": `

	$pkgName$+Service.on('$package$+_$evtName', (event: events.Event$EvtName) => {
	});
`,
}
