package goclienttemplates

var eventsGo = map[string]string{
	// *******************************
	"events.go": `
$#emit clientHeader

var $pkgName$+Handlers = map[string]func(*$PkgName$+Events, []string) {
$#each events eventHandler
}

type $PkgName$+Events struct {
$#each events eventHandlerMember
}

func (h *$PkgName$+Events) CallHandler(topic string, params []string) {
	handler := $pkgName$+Handlers[topic]
	if handler != nil {
		handler(h, params)
	}
}
$#each events funcSignature
$#each events eventClass
`,
	// *******************************
	"eventHandlerMember": `
	$evtName func(e *Event$EvtName)
`,
	// *******************************
	"funcSignature": `

func (h *$PkgName$+Events) On$PkgName$EvtName(handler func(e *Event$EvtName)) {
	h.$evtName = handler
}
`,
	// *******************************
	"eventHandler": `
	"$package.$evtName": func(evt *$PkgName$+Events, msg []string) { evt.on$PkgName$EvtName$+Thunk(msg) },
`,
	// *******************************
	"eventClass": `

type Event$EvtName struct {
	wasmclient.Event
$#each event eventClassField
}

func (h *$PkgName$+Events) on$PkgName$EvtName$+Thunk(message []string) {
    if h.$evtName == nil {
		return
	}
	e := &Event$EvtName{}
	e.Init(message)
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
	e.$FldName = e.Next$FldType()
`,
}
