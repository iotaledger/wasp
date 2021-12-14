package tsclienttemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `
import * as client from "wasmlib/client"
import * as app from "./$package"

export const eventHandlers: client.EventHandlers = {
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

export class Event$EvtName extends client.Event {
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
  	public $fldName: client.$FldType;
`,
	// *******************************
	"eventHandlerField": `
	e.$fldName = e.next$FldType();
`,
}
