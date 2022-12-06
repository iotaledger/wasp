// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var eventhandlersGo = map[string]string{
	// *******************************
	"eventhandlers.go": `
$#emit goHeader
$#emit importWasmTypes

var $pkgName$+Handlers = map[string]func(*$PkgName$+EventHandlers, []string) {
$#each events eventHandler
}

type $PkgName$+EventHandlers struct {
$#each events eventHandlerMember
}

var _ wasmlib.IEventHandlers = new($PkgName$+EventHandlers)

func (h *$PkgName$+EventHandlers) CallHandler(topic string, params []string) {
	handler := $pkgName$+Handlers[topic]
	if handler != nil {
		handler(h, params)
	}
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
	"$package.$evtName": func(evt *$PkgName$+EventHandlers, msg []string) { evt.on$PkgName$EvtName$+Thunk(msg) },
`,
	// *******************************
	"eventClass": `

type Event$EvtName struct {
	Timestamp uint64
$#each event eventClassField
}

func (h *$PkgName$+EventHandlers) on$PkgName$EvtName$+Thunk(msg []string) {
	if h.$evtName == nil {
		return
	}
	evt := wasmlib.NewEventDecoder(msg)
	e := &Event$EvtName{ Timestamp: evt.Timestamp() }
$#each event eventHandlerField
	h.$evtName(e)
}
`,
	// *******************************
	"eventClassField": `
	$FldName $fldLangType
`,
	// *******************************
	"eventHandlerField": `
	e.$FldName = wasmtypes.$FldType$+FromString(evt.Decode())
`,
}
