package goclienttemplates

var eventsGo = map[string]string{
	// *******************************
	"events.go": `
$#emit clientHeader

var EventHandlers = map[string]func([]string) {
$#each events eventHandler
}
$#each events eventClass
`,
	// *******************************
	"eventHandler": `
	"$package.$evtName": on$PkgName$EvtName$+Thunk,
`,
	// *******************************
	"eventClass": `

type Event$EvtName struct {
	wasmclient.Event
$#each event eventClassField
}

func on$PkgName$EvtName$+Thunk(message []string) {
	e := &Event$EvtName{}
	e.Init(message)
$#each event eventHandlerField
	On$PkgName$EvtName(e)
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
