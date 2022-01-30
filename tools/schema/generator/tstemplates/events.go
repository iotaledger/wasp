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
$#set params 
$#set separator 
$#each event eventParam

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
