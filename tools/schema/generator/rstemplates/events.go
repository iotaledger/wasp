package rstemplates

var eventsRs = map[string]string{
	// *******************************
	"events.rs": `
#![allow(dead_code)]

use wasmlib::*;

$#set TypeName $Package$+Events
pub struct $TypeName {
}

impl $TypeName {
$#each events eventFunc
}
`,
	// *******************************
	"eventFunc": `
$#set params 
$#each event eventParam

	pub fn $evt_name(&self$params) {
$#if event eventParams eventParamNone
	}
`,
	// *******************************
	"eventParam": `
$#set params $params, $fld_name: $fldRef$fldParamLangType
`,
	// *******************************
	"eventParamNone": `
		EventEncoder::new("$package.$evtName").emit();
`,
	// *******************************
	"eventParams": `
		let mut encoder = EventEncoder::new("$package.$evtName");
$#each event eventEmit
		encoder.emit();
`,
	// *******************************
	"eventEmit": `
		encoder.$fld_type($fldRef$fld_name);
`,
}
