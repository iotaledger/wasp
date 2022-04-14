// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package goclienttemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "Go Client",
	"extension":  ".go",
	"rootFolder": "go",
	"funcRegexp": `^func On(\w+).+$`,
}

var Templates = []map[string]string{
	config, // always first one
	common,
	eventsGo,
	serviceGo,
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
	"argEncode": {
		"Address":   "wasmtypes.ScBase58Decode",
		"AgentID":   "wasmtypes.ScBase58Decode",
		"BigInt":    "codec.EncodeBigInt",
		"Bool":      "codec.EncodeBool",
		"Bytes":     "[]byte",
		"ChainID":   "wasmtypes.ScBase58Decode",
		"Hash":      "wasmtypes.ScBase58Decode",
		"Hname":     "wasmtypes.ScBase58Decode",
		"Int8":      "codec.EncodeInt8",
		"Int16":     "codec.EncodeInt16",
		"Int32":     "codec.EncodeInt32",
		"Int64":     "codec.EncodeInt64",
		"NftID":     "wasmtypes.ScBase58Decode",
		"RequestID": "wasmtypes.ScBase58Decode",
		"String":    "codec.EncodeString",
		"TokenID":   "wasmtypes.ScBase58Decode",
		"Uint8":     "codec.EncodeUint8",
		"Uint16":    "codec.EncodeUint16",
		"Uint32":    "codec.EncodeUint32",
		"Uint64":    "codec.EncodeUint64",
	},
	"argDecode": {
		"Address":   "wasmtypes.ScBase58Encode",
		"AgentID":   "wasmtypes.ScBase58Encode",
		"BigInt":    "codec.DecodeBigInt",
		"Bool":      "codec.DecodeBool",
		"Bytes":     "[]byte",
		"ChainID":   "wasmtypes.ScBase58Encode",
		"Hash":      "wasmtypes.ScBase58Encode",
		"Hname":     "wasmtypes.ScBase58Encode",
		"Int8":      "codec.DecodeInt8",
		"Int16":     "codec.DecodeInt16",
		"Int32":     "codec.DecodeInt32",
		"Int64":     "codec.DecodeInt64",
		"NftID":     "wasmtypes.ScBase58Encode",
		"RequestID": "wasmtypes.ScBase58Encode",
		"String":    "codec.DecodeString",
		"TokenID":   "wasmtypes.ScBase58Encode",
		"Uint8":     "codec.DecodeUint8",
		"Uint16":    "codec.DecodeUint16",
		"Uint32":    "codec.DecodeUint32",
		"Uint64":    "codec.DecodeUint64",
	},
	"msgConvert": {
		"Address":   "e.Next()",
		"AgentID":   "e.Next()",
		"BigInt":    "e.NextBigInt()",
		"Bool":      "e.NextBool()",
		"ChainID":   "e.Next()",
		"Hash":      "e.Next()",
		"Hname":     "e.Next()",
		"Int8":      "e.NextInt8()",
		"Int16":     "e.NextInt16()",
		"Int32":     "e.NextInt32()",
		"Int64":     "e.NextInt64()",
		"NftID":     "e.Next()",
		"RequestID": "e.Next()",
		"String":    "e.Next()",
		"TokenID":   "e.Next()",
		"Uint8":     "e.NextUint8()",
		"Uint16":    "e.NextUint16()",
		"Uint32":    "e.NextUint32()",
		"Uint64":    "e.NextUint64()",
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
	"clientHeader": `
package $package$+client

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)
`,
}
