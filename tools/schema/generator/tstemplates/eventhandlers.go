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
    private $pkgName$+Handlers: Map<string, (evt: $PkgName$+EventHandlers, dec: wasmlib.WasmDecoder) => void> = new Map();

    /* eslint-disable @typescript-eslint/no-empty-function */
$#each events eventHandlerMember
    /* eslint-enable @typescript-eslint/no-empty-function */

    public constructor() {
        this.myID = wasmlib.eventHandlersGenerateID();
$#each events eventHandler
    }

    public callHandler(topic: string, dec: wasmlib.WasmDecoder): void {
        const handler = this.$pkgName$+Handlers.get(topic);
        if (handler) {
            handler(this, dec);
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
        this.$pkgName$+Handlers.set('$package.$evtName', (evt: $PkgName$+EventHandlers, dec: wasmlib.WasmDecoder) => evt.$evtName(new Event$EvtName(dec)));
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

    public constructor(dec: wasmlib.WasmDecoder) {
        this.timestamp = wasmtypes.uint64Decode(dec);
$#each event eventHandlerField
        dec.close();
    }
}
`,
	// *******************************
	"eventClassField": `
    public readonly $fldName: $fldLangType;
`,
	// *******************************
	"eventHandlerField": `
        this.$fldName = wasmtypes.$fldType$+Decode(dec);
`,
}
