package clienttemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "Client",
	"extension":  ".ts",
	"rootFolder": "client",
	"funcRegexp": `N/A`,
}

var Templates = []map[string]string{
	config,
	common,
	appTs,
	eventsTs,
	serviceTs,
}

var TypeDependent = model.StringMapMap{
	"msgConvert": {
		"Address":   "message[++index]",
		"AgentID":   "message[++index]",
		"Bool":      "message[++index][0]!='0'",
		"ChainID":   "message[++index]",
		"Color":     "message[++index]",
		"Hash":      "message[++index]",
		"Hname":     "message[++index]",
		"Int8":      "Number(message[++index])",
		"Int16":     "Number(message[++index])",
		"Int32":     "Number(message[++index])",
		"Int64":     "BigInt(message[++index])",
		"RequestID": "message[++index]",
		"String":    "message[++index]",
		"Uint8":     "Number(message[++index])",
		"Uint16":    "Number(message[++index])",
		"Uint32":    "Number(message[++index])",
		"Uint64":    "BigInt(message[++index])",
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
		"Address":   "toString",
		"AgentID":   "toString",
		"Bool":      "readUInt8",
		"ChainID":   "toString",
		"Color":     "toString",
		"Hash":      "toString",
		"Hname":     "toString",
		"Int8":      "readInt8",
		"Int16":     "readInt16LE",
		"Int32":     "readInt32LE",
		"Int64":     "readBigInt64LE",
		"RequestID": "toString",
		"String":    "toString",
		"Uint8":     "readUInt8",
		"Uint16":    "readUInt16LE",
		"Uint32":    "readUInt32LE",
		"Uint64":    "readBigUInt64LE",
	},
	"resConvert2": {
		"Bool": "!=0",
	},
}

var common = map[string]string{
	// *******************************
	"tmp": `
tmp`,
}
