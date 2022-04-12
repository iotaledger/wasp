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
$#set separator 
$#set params 
$#each event eventParam

$#if eventComment _eventComment
	$evtName($params): void {
		const evt = new wasmlib.EventEncoder("$package.$evtName");
$#each event eventEmit
		evt.emit();
	}
`,
	// *******************************
	"eventParam": `
$#set params $params$separator$fldName: $fldLangType
$#set separator , 
`,
	// *******************************
	"eventEmit": `
		evt.encode(wasmtypes.$fldType$+ToString($fldName));
`,
}
