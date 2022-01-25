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
	keysGo,
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
		"Bool":      "bool",
		"Bytes":     "[]byte",
		"ChainID":   "wasmtypes.ScChainID",
		"Color":     "wasmtypes.ScColor",
		"Hash":      "wasmtypes.ScHash",
		"Hname":     "wasmtypes.ScHname",
		"Int8":      "int8",
		"Int16":     "int16",
		"Int32":     "int32",
		"Int64":     "int64",
		"RequestID": "wasmtypes.ScRequestID",
		"String":    "string",
		"Uint8":     "uint8",
		"Uint16":    "uint16",
		"Uint32":    "uint32",
		"Uint64":    "uint64",
	},
	"fldToBytes": {
		"Address":   "key.Bytes()",
		"AgentID":   "key.Bytes()",
		"Bool":      "wasmtypes.BytesFromBool(key)",
		"Bytes":     "wasmtypes.BytesFromBytes(key)",
		"ChainID":   "key.Bytes()",
		"Color":     "key.Bytes()",
		"Hash":      "key.Bytes()",
		"Hname":     "key.Bytes()",
		"Int8":      "wasmtypes.BytesFromInt8(key)",
		"Int16":     "wasmtypes.BytesFromInt16(key)",
		"Int32":     "wasmtypes.BytesFromInt32(key)",
		"Int64":     "wasmtypes.BytesFromInt64(key)",
		"RequestID": "key.Bytes()",
		"String":    "wasmtypes.BytesFromString(key)",
		"Uint8":     "wasmtypes.BytesFromUint8(key)",
		"Uint16":    "wasmtypes.BytesFromUint16(key)",
		"Uint32":    "wasmtypes.BytesFromUint32(key)",
		"Uint64":    "wasmtypes.BytesFromUint64(key)",
		"":          "key.Bytes()",
	},
}

var common = map[string]string{
	// *******************************
	"goPackage": `
package $package
`,
	// *******************************
	"importWasmCodec": `
import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
`,
	// *******************************
	"importWasmLib": `
import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
`,
	// *******************************
	"importWasmTypes": `
import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
`,
	// *******************************
	"goHeader": `
$#emit goPackage

$#emit importWasmLib
`,
}
