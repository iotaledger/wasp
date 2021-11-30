package gotemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var Templates = []map[string]string{
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
		"Address":   "wasmlib.ScAddress",
		"AgentID":   "wasmlib.ScAgentID",
		"Bool":      "bool",
		"ChainID":   "wasmlib.ScChainID",
		"Color":     "wasmlib.ScColor",
		"Hash":      "wasmlib.ScHash",
		"Hname":     "wasmlib.ScHname",
		"Int8":      "int8",
		"Int16":     "int16",
		"Int32":     "int32",
		"Int64":     "int64",
		"RequestID": "wasmlib.ScRequestID",
		"String":    "string",
		"Uint8":     "uint8",
		"Uint16":    "uint16",
		"Uint32":    "uint32",
		"Uint64":    "uint64",
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
		"Address":   "key",
		"AgentID":   "key",
		"Bool":      "wasmlib.Key32(key)",
		"ChainID":   "key",
		"Color":     "key",
		"Hash":      "key",
		"Hname":     "key",
		"Int8":      "wasmlib.Key32(key)",
		"Int16":     "wasmlib.Key32(key)",
		"Int32":     "wasmlib.Key32(key)",
		"Int64":     "wasmlib.Key32(key)",
		"RequestID": "key",
		"String":    "wasmlib.Key(key)",
		"Uint8":     "wasmlib.Key32(key)",
		"Uint16":    "wasmlib.Key32(key)",
		"Uint32":    "wasmlib.Key32(key)",
		"Uint64":    "wasmlib.Key32(key)",
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
	"goPackage": `
package $package
`,
	// *******************************
	"importWasmLib": `

import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
`,
	// *******************************
	"goHeader": `
$#emit goPackage
$#emit importWasmLib
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
