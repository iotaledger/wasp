// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "Rust",
	"extension":  ".rs",
	"rootFolder": "src",
	"funcRegexp": `^pub fn (\w+).+$`,
}

var Templates = []map[string]string{
	config, // always first one
	common,
	cargoToml,
	constsRs,
	contractRs,
	eventsRs,
	funcsRs,
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
		"Bytes":     "Vec<u8>",
		"ChainID":   "ScChainID",
		"Color":     "ScTokenID",
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
	"fldParamLangType": {
		"Address":   "ScAddress",
		"AgentID":   "ScAgentID",
		"Bool":      "bool",
		"Bytes":     "[u8]",
		"ChainID":   "ScChainID",
		"Color":     "ScTokenID",
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
		"Bytes":     "&",
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
$#set crate 
$#if core setCrate
`,
	// *******************************
	"setCrate": `
$#set crate (crate)
`,
	// *******************************
	"useCoreContract": `
use crate::$package::*;
`,
	// *******************************
	"useCrate": `
use crate::*;
`,
	// *******************************
	"useWasmLib": `
use wasmlib::*;
`,
}
