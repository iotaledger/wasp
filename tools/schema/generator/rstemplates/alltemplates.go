package rstemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var Templates = []map[string]string{
	common,
	cargoToml,
	constsRs,
	contractRs,
	eventsRs,
	funcsRs,
	keysRs,
	libRs,
	modRs,
	paramsRs,
	proxyRs,
	resultsRs,
	stateRs,
	structsRs,
	typedefsRs,
}

var TypeDependent = model.StringMapMap{
	"fldLangType": {
		"Address":   "ScAddress",
		"AgentID":   "ScAgentID",
		"Bool":      "bool",
		"ChainID":   "ScChainID",
		"Color":     "ScColor",
		"Hash":      "ScHash",
		"Hname":     "ScHname",
		"Int8":      "i8",
		"Int16":     "i16",
		"Int32":     "i32",
		"Int64":     "i64",
		"RequestID": "ScRequestID",
		"String":    "String",
		"Uint8":     "u8",
		"Uint16":    "u16",
		"Uint32":    "u32",
		"Uint64":    "u64",
	},
	"fldTypeID": {
		"Address":   "TYPE_ADDRESS",
		"AgentID":   "TYPE_AGENT_ID",
		"Bool":      "TYPE_BOOL",
		"ChainID":   "TYPE_CHAIN_ID",
		"Color":     "TYPE_COLOR",
		"Hash":      "TYPE_HASH",
		"Hname":     "TYPE_HNAME",
		"Int8":      "TYPE_INT8",
		"Int16":     "TYPE_INT16",
		"Int32":     "TYPE_INT32",
		"Int64":     "TYPE_INT64",
		"RequestID": "TYPE_REQUEST_ID",
		"String":    "TYPE_STRING",
		"Uint8":     "TYPE_INT8",
		"Uint16":    "TYPE_INT16",
		"Uint32":    "TYPE_INT32",
		"Uint64":    "TYPE_INT64",
		"":          "TYPE_BYTES",
	},
	"fldToKey32": {
		"Address":   "key",
		"AgentID":   "key",
		"Bool":      "Key32(key)",
		"ChainID":   "key",
		"Color":     "key",
		"Hash":      "key",
		"Hname":     "key",
		"Int8":      "Key32(key)",
		"Int16":     "Key32(key)",
		"Int32":     "Key32(key)",
		"Int64":     "Key32(key)",
		"RequestID": "key",
		"String":    "key",
		"Uint8":     "Key32(key)",
		"Uint16":    "Key32(key)",
		"Uint32":    "Key32(key)",
		"Uint64":    "Key32(key)",
	},
	"fldParamLangType": {
		"Address":   "ScAddress",
		"AgentID":   "ScAgentID",
		"Bool":      "bool",
		"ChainID":   "ScChainID",
		"Color":     "ScColor",
		"Hash":      "ScHash",
		"Hname":     "ScHname",
		"Int8":      "i8",
		"Int16":     "i16",
		"Int32":     "i32",
		"Int64":     "i64",
		"RequestID": "ScRequestID",
		"String":    "str",
		"Uint8":     "u8",
		"Uint16":    "u16",
		"Uint32":    "u32",
		"Uint64":    "u64",
	},
	"fldRef": {
		"Address":   "&",
		"AgentID":   "&",
		"ChainID":   "&",
		"Color":     "&",
		"Hash":      "&",
		"RequestID": "&",
		"String":    "&",
	},
}

var common = map[string]string{
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
	"rsHeader": `
$#if core useCrate useWasmLib
`,
	// *******************************
	"modEvents": `
mod events;
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
	"useEvents": `
use crate::events::*;
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
	// *******************************
	"setVarType": `
$#set varType TYPE_MAP
$#if array setVarTypeArray
`,
	// *******************************
	"setVarTypeArray": `
$#set varType $arrayTypeID | $fldTypeID
`,
}
