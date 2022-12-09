// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var eventsGo = map[string]string{
	// *******************************
	"events.go": `
//nolint:gocritic
package $package

$#emit importWasmLibAndWasmTypes

$#set TypeName $Package$+Events
type $TypeName struct {
}
$#each events eventFunc
`,
	// *******************************
	"eventFunc": `
$#set endFunc ) {
$#if event eventSetEndFunc

$#each eventComment _eventComment
func (e $TypeName) $EvtName($endFunc
$#each event eventParam
$#if event eventEndFunc2
	evt := wasmlib.NewEventEncoder("$package.$evtName")
$#each event eventEmit
	evt.Emit()
}
`,
	// *******************************
	"eventParam": `
$#each fldComment _eventParamComment
	$fldName $fldLangType,
`,
	// *******************************
	"eventEmit": `
	evt.Encode(wasmtypes.$FldType$+ToString($fldName))
`,
	// *******************************
	"eventSetEndFunc": `
$#set endFunc $nil
`,
	// *******************************
	"eventEndFunc2": `
) {
`,
}
