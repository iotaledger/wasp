package tsclienttemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `
$#emit importWasmLib
import * as app from "./$package"

export const eventHandlers: wasmclient.EventHandlers = new Map([
$#each events eventHandler
]);
$#each events eventClass
`,
	// *******************************
	"eventHandler": `
	["$package.$evtName", (msg: string[]) => app.on$PkgName$EvtName(new Event$EvtName(msg))],
`,
	// *******************************
	"eventClass": `

export class Event$EvtName extends wasmclient.Event {
$#each event eventClassField
	
	public constructor(msg: string[]) {
		super(msg)
$#each event eventHandlerField
	}
}
`,
	// *******************************
	"eventClassField": `
	public readonly $fldName: wasmclient.$FldType;
`,
	// *******************************
	"eventHandlerField": `
		this.$fldName = this.next$FldType();
`,
}
