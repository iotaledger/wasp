// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var eventhandlersTs = map[string]string{
	// *******************************
	"eventhandlers.ts": `
$#emit importWasmLib
$#emit importWasmTypes

export class $PkgName$+EventHandlers implements wasmlib.IEventHandlers {
    private myID: u32;
    private $pkgName$+Handlers: Map<string, (evt: $PkgName$+EventHandlers, msg: string[]) => void> = new Map();

    /* eslint-disable @typescript-eslint/no-empty-function */
$#each events eventHandlerMember
    /* eslint-enable @typescript-eslint/no-empty-function */

    public constructor() {
        this.myID = wasmlib.eventHandlersGenerateID();
$#each events eventHandler
    }

    public callHandler(topic: string, params: string[]): void {
        const handler = this.$pkgName$+Handlers.get(topic);
        if (handler) {
            handler(this, params);
        }
    }

    public id(): u32 {
        return this.myID;
    }
$#each events eventFuncSignature
}
$#each events eventClass
`,
	// *******************************
	"eventHandler": `
        this.$pkgName$+Handlers.set('$package.$evtName', (evt: $PkgName$+EventHandlers, msg: string[]) => evt.$evtName(new Event$EvtName(msg)));
`,
	// *******************************
	"eventHandlerMember": `
    $evtName: (evt: Event$EvtName) => void = () => {};
`,
	// *******************************
	"eventFuncSignature": `

    public on$PkgName$EvtName(handler: (evt: Event$EvtName) => void): void {
        this.$evtName = handler;
    }
`,
	// *******************************
	"eventClass": `

export class Event$EvtName {
    public readonly timestamp: u64;
$#each event eventClassField

    public constructor(msg: string[]) {
        const evt = new wasmlib.EventDecoder(msg);
        this.timestamp = evt.timestamp();
$#each event eventHandlerField
    }
}
`,
	// *******************************
	"eventClassField": `
    public readonly $fldName: $fldLangType;
`,
	// *******************************
	"eventHandlerField": `
        this.$fldName = wasmtypes.$fldType$+FromString(evt.decode());
`,
}
