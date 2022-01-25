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
$#set params 
$#set separator 
$#each event eventParam

func (e $TypeName) $EvtName($params) {
	wasmlib.NewEventEncoder("$package.$evtName").
$#each event eventEmit
		Emit()
}
`,
	// *******************************
	"eventParam": `
$#set params $params$separator$fldName $fldLangType
$#set separator , 
`,
	// *******************************
	"eventEmit": `
		$FldType($fldName).
`,
}
