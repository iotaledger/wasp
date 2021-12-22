package tsclienttemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "TypeScript Client",
	"extension":  ".ts",
	"rootFolder": "client/ts",
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
		"Address":   "client.Address",
		"AgentID":   "client.AgentID",
		"Bool":      "boolean",
		"Bytes":     "client.Bytes",
		"ChainID":   "client.ChainID",
		"Color":     "client.Color",
		"Hash":      "client.Hash",
		"Hname":     "client.Hname",
		"Int8":      "client.Int8",
		"Int16":     "client.Int16",
		"Int32":     "client.Int32",
		"Int64":     "client.Int64",
		"RequestID": "client.RequestID",
		"String":    "string",
		"Uint8":     "client.Uint8",
		"Uint16":    "client.Uint16",
		"Uint32":    "client.Uint32",
		"Uint64":    "client.Uint64",
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
}
