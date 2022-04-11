// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var eventsGo = map[string]string{
	// *******************************
	"events.go": `
//nolint:gocritic
$#emit goHeader
$#emit importWasmTypes

$#set TypeName $Package$+Events
type $TypeName struct {
}
$#each events eventFunc
`,
	// *******************************
	"eventFunc": `
$#set separator 
$#set params 
$#each event eventParam

$#if funcComment _eventComment
func (e $TypeName) $EvtName(
$params) {
	evt := wasmlib.NewEventEncoder("$package.$evtName")
$#each event eventEmit
	evt.Emit()
}
`,
	// *******************************
	"eventParam": "\n$#set params $params$fldName $fldLangType, $fldComment\r\n",
	// *******************************
	"eventEmit": `
	evt.Encode(wasmtypes.$FldType$+ToString($fldName))
`,
	// *******************************
	"_eventComment": `
$eventComment
`,
}
