package tsclienttemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `
import * as wasmclient from "./wasmclient"
import * as app from "./$package"

export const eventHandlers: wasmclient.EventHandlers = {
$#each events eventHandler
};
$#each events eventClass
`,
	// *******************************
	"eventHandler": `
	"$package.$evtName": msg => app.on$PkgName$EvtName(new Event$EvtName(msg)),
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
