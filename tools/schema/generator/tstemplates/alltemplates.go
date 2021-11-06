package tstemplates

var TsTemplates = []map[string]string{
	tsCommon,
	constsTs,
	contractTs,
	funcsTs,
	keysTs,
	libTs,
	paramsTs,
	proxyTs,
	resultsTs,
	stateTs,
	structsTs,
	typedefsTs,
}

var tsCommon = map[string]string{
	// *******************************
	"initGlobals": `
$#set arrayTypeID TYPE_ARRAY
$#set crate 
$#if core setArrayTypeID
`,
	// *******************************
	"setArrayTypeID": `
$#set arrayTypeID TYPE_ARRAY16
$#set crate (crate)
`,
	// *******************************
	"tsHeader": `
$#if core useCrate useWasmLib
`,
	// *******************************
	"modParams": `
mod params;
`,
	// *******************************
	"modResults": `
mod results;
`,
	// *******************************
	"modStructs": `
mod structs;
`,
	// *******************************
	"modTypeDefs": `
mod typedefs;
`,
	// *******************************
	"useCrate": `
use crate::*;
`,
	// *******************************
	"useCoreContract": `
use crate::$package::*;
`,
	// *******************************
	"useHost": `
use crate::host::*;
`,
	// *******************************
	"useParams": `
use crate::params::*;
`,
	// *******************************
	"useResults": `
use crate::results::*;
`,
	// *******************************
	"useStructs": `
use crate::structs::*;
`,
	// *******************************
	"useTypeDefs": `
use crate::typedefs::*;
`,
	// *******************************
	"useWasmLib": `
use wasmlib::*;
`,
}
