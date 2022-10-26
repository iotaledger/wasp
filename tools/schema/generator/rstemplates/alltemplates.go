// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "Rust",
	"extension":  ".rs",
	"rootFolder": "rs",
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
	mainRs,
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
		"BigInt":    "ScBigInt",
		"Bool":      "bool",
		"Bytes":     "Vec<u8>",
		"ChainID":   "ScChainID",
		"Hash":      "ScHash",
		"Hname":     "ScHname",
		"Int8":      "i8",
		"Int16":     "i16",
		"Int32":     "i32",
		"Int64":     "i64",
		"NftID":     "ScNftID",
		"RequestID": "ScRequestID",
		"String":    "String",
		"TokenID":   "ScTokenID",
		"Uint8":     "u8",
		"Uint16":    "u16",
		"Uint32":    "u32",
		"Uint64":    "u64",
	},
	"fldParamLangType": {
		"Address":   "ScAddress",
		"AgentID":   "ScAgentID",
		"BigInt":    "ScBigInt",
		"Bool":      "bool",
		"Bytes":     "[u8]",
		"ChainID":   "ScChainID",
		"Hash":      "ScHash",
		"Hname":     "ScHname",
		"Int8":      "i8",
		"Int16":     "i16",
		"Int32":     "i32",
		"Int64":     "i64",
		"NftID":     "ScNftID",
		"RequestID": "ScRequestID",
		"String":    "str",
		"TokenID":   "ScTokenID",
		"Uint8":     "u8",
		"Uint16":    "u16",
		"Uint32":    "u32",
		"Uint64":    "u64",
	},
	"fldRef": {
		"Address":   "&",
		"AgentID":   "&",
		"BigInt":    "&",
		"Bytes":     "&",
		"ChainID":   "&",
		"Hash":      "&",
		"NftID":     "&",
		"RequestID": "&",
		"String":    "&",
		"TokenID":   "&",
	},
}

var common = map[string]string{
	// *******************************
	"initGlobals": `
$#set crate $nil
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
	// *******************************
	"../LICENSE": `
https://www.apache.org/licenses/LICENSE-2.0
`,
	// *******************************
	"_eventComment": `
    $nextLine
`,
	// *******************************
	"_eventParamComment": `
        $nextLine
`,
	// *******************************
	"_fldComment": `
    $nextLine
`,
	// *******************************
	"_funcComment": `
    $nextLine
`,
	// *******************************
	"_funcAccessComment": `
    $nextLine
`,
	// *******************************
	"_structComment": `
$nextLine
`,
	// *******************************
	"_structFieldComment": `
    $nextLine
`,
	// *******************************
	"_typedefComment": `
$nextLine
`,
}
