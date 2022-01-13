package tsclienttemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `
$#emit importWasmLib

type $PkgName$+Handlers = Map<string, (evt: $PkgName$+Events, msg: string[]) => void>;

export class $PkgName$+Events implements wasmclient.IEventHandler {
	private eventHandlers: $PkgName$+Handlers = new Map([
$#each events eventHandler
	]);

	public callHandler(topic: string, params: string[]): void {
		const handler = this.eventHandlers.get(topic);
		if (handler !== undefined) {
			handler(this, params);
		}
	}
$#each events funcSignature
}
$#each events eventClass
`,
	// *******************************
	"eventHandler": `
		["$package.$evtName", (evt: $PkgName$+Events, msg: string[]) => evt.on$PkgName$EvtName(new Event$EvtName(msg))],
`,
	// *******************************
	"funcSignature": `

	public on$PkgName$EvtName(event: Event$EvtName): void {
	}
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
