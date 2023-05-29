// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var eventhandlersGo = map[string]string{
	// *******************************
	"eventhandlers.go": `
package $package

$#emit importWasmLibAndWasmTypes

var $pkgName$+Handlers = map[string]func(*$PkgName$+EventHandlers, *wasmtypes.WasmDecoder){
$#each events eventHandler
}

type $PkgName$+EventHandlers struct {
	myID uint32
$#each events eventHandlerMember
}

var _ wasmlib.IEventHandlers = new($PkgName$+EventHandlers)

func New$PkgName$+EventHandlers() *$PkgName$+EventHandlers {
	return &$PkgName$+EventHandlers{ myID: wasmlib.EventHandlersGenerateID() }
}

func (h *$PkgName$+EventHandlers) CallHandler(topic string, dec *wasmtypes.WasmDecoder) {
	handler := $pkgName$+Handlers[topic]
	if handler != nil {
		handler(h, dec)
	}
}

func (h *$PkgName$+EventHandlers) ID() uint32 {
	return h.myID
}
$#each events eventFuncSignature
$#each events eventClass
`,
	// *******************************
	"eventHandlerMember": `
	$evtName func(e *Event$EvtName)
`,
	// *******************************
	"eventFuncSignature": `

func (h *$PkgName$+EventHandlers) On$PkgName$EvtName(handler func(e *Event$EvtName)) {
	h.$evtName = handler
}
`,
	// *******************************
	"eventHandler": `
	"$package.$evtName": func(evt *$PkgName$+EventHandlers, dec *wasmtypes.WasmDecoder) { evt.on$PkgName$EvtName$+Thunk(dec) },
`,
	// *******************************
	"eventClass": `

type Event$EvtName struct {
	Timestamp uint64
$#each event eventClassField
}

func (h *$PkgName$+EventHandlers) on$PkgName$EvtName$+Thunk(dec *wasmtypes.WasmDecoder) {
	if h.$evtName == nil {
		return
	}
	e := &Event$EvtName{}
	e.Timestamp = wasmtypes.Uint64Decode(dec)
$#each event eventHandlerField
	dec.Close()
	h.$evtName(e)
}
`,
	// *******************************
	"eventClassField": `
	$FldName $fldLangType
`,
	// *******************************
	"eventHandlerField": `
	e.$FldName = wasmtypes.$FldType$+Decode(dec)
`,
}
