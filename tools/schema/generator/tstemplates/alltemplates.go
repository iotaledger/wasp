package tstemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "TypeScript",
	"extension":  ".ts",
	"rootFolder": "ts",
	"funcRegexp": `^export function (\w+).+$`,
}

var Templates = []map[string]string{
	config, // always first one
	common,
	constsTs,
	contractTs,
	eventsTs,
	funcsTs,
	indexTs,
	keysTs,
	libTs,
	paramsTs,
	proxyTs,
	resultsTs,
	stateTs,
	structsTs,
	typedefsTs,
}

var TypeDependent = model.StringMapMap{
	"fldLangType": {
		"Address":   "wasmlib.ScAddress",
		"AgentID":   "wasmlib.ScAgentID",
		"Bool":      "bool",
		"ChainID":   "wasmlib.ScChainID",
		"Color":     "wasmlib.ScColor",
		"Hash":      "wasmlib.ScHash",
		"Hname":     "wasmlib.ScHname",
		"Int8":      "i8",
		"Int16":     "i16",
		"Int32":     "i32",
		"Int64":     "i64",
		"RequestID": "wasmlib.ScRequestID",
		"String":    "string",
		"Uint8":     "u8",
		"Uint16":    "u16",
		"Uint32":    "u32",
		"Uint64":    "u64",
	},
	"fldTypeID": {
		"Address":   "wasmlib.TYPE_ADDRESS",
		"AgentID":   "wasmlib.TYPE_AGENT_ID",
		"Bool":      "wasmlib.TYPE_BOOL",
		"ChainID":   "wasmlib.TYPE_CHAIN_ID",
		"Color":     "wasmlib.TYPE_COLOR",
		"Hash":      "wasmlib.TYPE_HASH",
		"Hname":     "wasmlib.TYPE_HNAME",
		"Int8":      "wasmlib.TYPE_INT8",
		"Int16":     "wasmlib.TYPE_INT16",
		"Int32":     "wasmlib.TYPE_INT32",
		"Int64":     "wasmlib.TYPE_INT64",
		"RequestID": "wasmlib.TYPE_REQUEST_ID",
		"String":    "wasmlib.TYPE_STRING",
		"Uint8":     "wasmlib.TYPE_INT8",
		"Uint16":    "wasmlib.TYPE_INT16",
		"Uint32":    "wasmlib.TYPE_INT32",
		"Uint64":    "wasmlib.TYPE_INT64",
		"":          "wasmlib.TYPE_BYTES",
	},
	"fldToKey32": {
		"Address":   "key.getKeyID()",
		"AgentID":   "key.getKeyID()",
		"Bool":      "???cannot use Bool as map key",
		"ChainID":   "key.getKeyID()",
		"Color":     "key.getKeyID()",
		"Hash":      "key.getKeyID()",
		"Hname":     "key.getKeyID()",
		"Int8":      "wasmlib.getKeyIDFromUint64(key as u64, 1)",
		"Int16":     "wasmlib.getKeyIDFromUint64(key as u64, 2)",
		"Int32":     "wasmlib.getKeyIDFromUint64(key as u64, 4)",
		"Int64":     "wasmlib.getKeyIDFromUint64(key as u64, 8)",
		"RequestID": "key.getKeyID()",
		"String":    "wasmlib.Key32.fromString(key)",
		"Uint8":     "wasmlib.getKeyIDFromUint64(key as u64, 1)",
		"Uint16":    "wasmlib.getKeyIDFromUint64(key as u64, 2)",
		"Uint32":    "wasmlib.getKeyIDFromUint64(key as u64, 4)",
		"Uint64":    "wasmlib.getKeyIDFromUint64(key, 8)",
	},
	"fldTypeInit": {
		"Address":   "new wasmlib.ScAddress()",
		"AgentID":   "new wasmlib.ScAgentID()",
		"Bool":      "false",
		"ChainID":   "new wasmlib.ScChainID()",
		"Color":     "new wasmlib.ScColor(0)",
		"Hash":      "new wasmlib.ScHash()",
		"Hname":     "new wasmlib.ScHname(0)",
		"Int8":      "0",
		"Int16":     "0",
		"Int32":     "0",
		"Int64":     "0",
		"RequestID": "new wasmlib.ScRequestID()",
		"String":    "\"\"",
		"Uint8":     "0",
		"Uint16":    "0",
		"Uint32":    "0",
		"Uint64":    "0",
	},
}

var common = map[string]string{
	// *******************************
	"initGlobals": `
$#set arrayTypeID wasmlib.TYPE_ARRAY
$#if core setArrayTypeID
`,
	// *******************************
	"setArrayTypeID": `
$#set arrayTypeID wasmlib.TYPE_ARRAY16
`,
	// *******************************
	"importWasmLib": `
import * as wasmlib from "wasmlib";
`,
	// *******************************
	"importSc": `
import * as sc from "./index";
`,
	// *******************************
	"tsImports": `
$#emit importWasmLib
$#emit importSc
`,
	// *******************************
	"tsconfig.json": `
{
  "extends": "assemblyscript/std/assembly.json",
  "include": ["./*.ts"]
}
`,
	// *******************************
	"setVarType": `
$#set varType wasmlib.TYPE_MAP
$#if array setVarTypeArray
`,
	// *******************************
	"setVarTypeArray": `
$#set varType $arrayTypeID|$fldTypeID
`,
}
