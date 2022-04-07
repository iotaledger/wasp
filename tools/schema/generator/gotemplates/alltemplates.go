// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "Go",
	"extension":  ".go",
	"rootFolder": "go",
	"funcRegexp": `^func (\w+).+$`,
}

var Templates = []map[string]string{
	config, // always first one
	common,
	constsGo,
	contractGo,
	eventsGo,
	funcsGo,
	libGo,
	mainGo,
	paramsGo,
	proxyGo,
	resultsGo,
	stateGo,
	structsGo,
	typedefsGo,
}

var TypeDependent = model.StringMapMap{
	"fldLangType": {
		"Address":   "wasmtypes.ScAddress",
		"AgentID":   "wasmtypes.ScAgentID",
		"BigInt":    "wasmtypes.ScBigInt",
		"Bool":      "bool",
		"Bytes":     "[]byte",
		"ChainID":   "wasmtypes.ScChainID",
		"Hash":      "wasmtypes.ScHash",
		"Hname":     "wasmtypes.ScHname",
		"Int8":      "int8",
		"Int16":     "int16",
		"Int32":     "int32",
		"Int64":     "int64",
		"NftID":     "wasmtypes.ScNftID",
		"RequestID": "wasmtypes.ScRequestID",
		"String":    "string",
		"TokenID":   "wasmtypes.ScTokenID",
		"Uint8":     "uint8",
		"Uint16":    "uint16",
		"Uint32":    "uint32",
		"Uint64":    "uint64",
	},
}

var common = map[string]string{
	// *******************************
	"importWasmLib": `
import "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
`,
	// *******************************
	"importWasmTypes": `
import "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
`,
	// *******************************
	"goPackage": `
package $package
`,
	// *******************************
	"goHeader": `
$#emit goPackage

$#emit importWasmLib
`,
	// *******************************
	"_fldComment": `
$fldComment
`,
}
