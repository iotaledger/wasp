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
        const enc = wasmlib.eventEncoder('$package.$evtName');
$#each event eventEmit
        wasmlib.eventEmit(enc);
    }
`,
	// *******************************
	"eventParam": `
$#each fldComment _eventParamComment
        $fldName: $fldLangType,
`,
	// *******************************
	"eventEmit": `
        wasmtypes.$fldType$+Encode(enc, $fldName);
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
