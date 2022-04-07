// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

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
	indexTs,
	serviceTs,
}

var TypeDependent = model.StringMapMap{
	"fldLangType": {
		"Address":   "wasmclient.Address",
		"AgentID":   "wasmclient.AgentID",
		"BigInt":    "BigInt",
		"Bool":      "boolean",
		"Bytes":     "wasmclient.Bytes",
		"ChainID":   "wasmclient.ChainID",
		"Hash":      "wasmclient.Hash",
		"Hname":     "wasmclient.Hname",
		"Int8":      "wasmclient.Int8",
		"Int16":     "wasmclient.Int16",
		"Int32":     "wasmclient.Int32",
		"Int64":     "wasmclient.Int64",
		"NftID":     "wasmtypes.ScNftID",
		"RequestID": "wasmclient.RequestID",
		"String":    "string",
		"TokenID":   "wasmclient.TokenID",
		"Uint8":     "wasmclient.Uint8",
		"Uint16":    "wasmclient.Uint16",
		"Uint32":    "wasmclient.Uint32",
		"Uint64":    "wasmclient.Uint64",
	},
	"fldDefault": {
		"Address":   "''",
		"AgentID":   "''",
		"BigInt":    "BigInt(0)",
		"Bool":      "false",
		"ChainID":   "''",
		"Hash":      "''",
		"Hname":     "''",
		"Int8":      "0",
		"Int16":     "0",
		"Int32":     "0",
		"Int64":     "BigInt(0)",
		"NftID":     "''",
		"RequestID": "''",
		"String":    "''",
		"TokenID":   "''",
		"Uint8":     "0",
		"Uint16":    "0",
		"Uint32":    "0",
		"Uint64":    "BigInt(0)",
	},
	"resConvert": {
		"Address":   "toString()",
		"AgentID":   "toString()",
		"BigInt":    "readBigUInt64LE(0)",
		"Bool":      "readUInt8(0)!=0",
		"ChainID":   "toString()",
		"Hash":      "toString()",
		"Hname":     "toString()",
		"Int8":      "readInt8(0)",
		"Int16":     "readInt16LE(0)",
		"Int32":     "readInt32LE(0)",
		"Int64":     "readBigInt64LE(0)",
		"NftID":     "toString()",
		"RequestID": "toString()",
		"String":    "toString()",
		"TokenID":   "toString()",
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
	"importWasmClient": `
import * as wasmclient from "wasmclient"
`,
}
