package tstemplates

var eventsTs = map[string]string{
	// *******************************
	"events.ts": `
$#emit importWasmLib

$#set TypeName $Package$+Events
export class $TypeName {
$#each events eventFunc
}
`,
	// *******************************
	"eventFunc": `
$#set params 
$#set separator 
$#each event eventParam

	$evtName($params): void {
		new wasmlib.EventEncoder("$package.$evtName").
$#each event eventEmit
		emit();
	}
`,
	// *******************************
	"eventParam": `
$#set params $params$separator$fldName: $fldLangType
$#set separator , 
`,
	// *******************************
	"eventEmit": `
		$fldType($fldName).
`,
}
