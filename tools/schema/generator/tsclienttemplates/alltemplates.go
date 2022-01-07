package tsclienttemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "TypeScript Client",
	"extension":  ".ts",
	"rootFolder": "ts",
	"funcRegexp": `^export function on(\w+).+$`,
}

var Templates = []map[string]string{
	config, // always first one
	common,
	eventsTs,
	funcsTs,
	indexTs,
	serviceTs,
}

var TypeDependent = model.StringMapMap{
	"fldLangType": {
		"Address":   "wasmclient.Address",
		"AgentID":   "wasmclient.AgentID",
		"Bool":      "boolean",
		"Bytes":     "wasmclient.Bytes",
		"ChainID":   "wasmclient.ChainID",
		"Color":     "wasmclient.Color",
		"Hash":      "wasmclient.Hash",
		"Hname":     "wasmclient.Hname",
		"Int8":      "wasmclient.Int8",
		"Int16":     "wasmclient.Int16",
		"Int32":     "wasmclient.Int32",
		"Int64":     "wasmclient.Int64",
		"RequestID": "wasmclient.RequestID",
		"String":    "string",
		"Uint8":     "wasmclient.Uint8",
		"Uint16":    "wasmclient.Uint16",
		"Uint32":    "wasmclient.Uint32",
		"Uint64":    "wasmclient.Uint64",
	},
	"fldDefault": {
		"Address":   "''",
		"AgentID":   "''",
		"Bool":      "false",
		"ChainID":   "''",
		"Color":     "''",
		"Hash":      "''",
		"Hname":     "''",
		"Int8":      "0",
		"Int16":     "0",
		"Int32":     "0",
		"Int64":     "BigInt(0)",
		"RequestID": "''",
		"String":    "''",
		"Uint8":     "0",
		"Uint16":    "0",
		"Uint32":    "0",
		"Uint64":    "BigInt(0)",
	},
	"resConvert": {
		"Address":   "toString()",
		"AgentID":   "toString()",
		"Bool":      "readUInt8(0)!=0",
		"ChainID":   "toString()",
		"Color":     "toString()",
		"Hash":      "toString()",
		"Hname":     "toString()",
		"Int8":      "readInt8(0)",
		"Int16":     "readInt16LE(0)",
		"Int32":     "readInt32LE(0)",
		"Int64":     "readBigInt64LE(0)",
		"RequestID": "toString()",
		"String":    "toString()",
		"Uint8":     "readUInt8(0)",
		"Uint16":    "readUInt16LE(0)",
		"Uint32":    "readUInt32LE(0)",
		"Uint64":    "readBigUInt64LE(0)",
	},
}

var common = map[string]string{
	// *******************************
	"tsconfig.json": `
{
  "compilerOptions": {
    "module": "commonjs",
    "lib": ["es2020"],
    "target": "es2020",
    "sourceMap": true
  },
  "exclude": [
    "node_modules"
  ],
}
`,
	// *******************************
	"importEvents": `
import * as events from "./events"
`,
	// *******************************
	"importService": `
import * as service from "./service"
`,
	// *******************************
	"importWasmLib": `
import * as wasmclient from "wasmclient"
`,
}
