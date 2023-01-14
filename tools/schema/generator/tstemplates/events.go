// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `
$#emit importWasmLib
$#emit importWasmTypes

$#set TypeName $Package$+Events
export class $TypeName {
$#each events eventFunc
}
`,
	// *******************************
	"eventFunc": `
$#set endFunc ): void {
$#if event eventSetEndFunc

$#each eventComment _eventComment
    $evtName($endFunc
$#each event eventParam
$#if event eventEndFunc2
        const evt = new wasmlib.EventEncoder('$package.$evtName');
$#each event eventEmit
        evt.emit();
    }
`,
	// *******************************
	"eventParam": `
$#each fldComment _eventParamComment
        $fldName: $fldLangType,
`,
	// *******************************
	"eventEmit": `
        evt.encode(wasmtypes.$fldType$+ToString($fldName));
`,
	// *******************************
	"eventSetEndFunc": `
$#set endFunc $nil
`,
	// *******************************
	"eventEndFunc2": `
    ): void {
`,
}
