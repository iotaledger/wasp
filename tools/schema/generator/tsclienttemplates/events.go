package tsclienttemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `
import * as wasmclient from "wasmlib/wasmclient"
import * as app from "./$package"

export const eventHandlers: wasmclient.EventHandlers = {
$#each events eventHandler
};
$#each events eventClass
`,
	// *******************************
	"eventHandler": `
	"$package.$evtName": on$PkgName$EvtName$+Thunk,
`,
	// *******************************
	"eventClass": `

export class Event$EvtName extends wasmclient.Event {
$#each event eventClassField
}

function on$PkgName$EvtName$+Thunk(message: string[]) {
	let e = new Event$EvtName(message);
$#each event eventHandlerField
	app.on$PkgName$EvtName(e);
}
`,
	// *******************************
	"eventClassField": `
  	public $fldName: wasmclient.$FldType;
`,
	// *******************************
	"eventHandlerField": `
	e.$fldName = e.next$FldType();
`,
}
