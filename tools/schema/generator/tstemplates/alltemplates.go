package tstemplates

var TsTemplates = []map[string]string{
	tsCommon,
	constsTs,
	contractTs,
	funcsTs,
	keysTs,
	libTs,
	paramsTs,
	proxyTs,
	resultsTs,
	stateTs,
	structsTs,
	typedefsTs,
}

var tsCommon = map[string]string{
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
	"importWasmLib": `
import * as wasmlib from "wasmlib";
`,
	// *******************************
	"importSc": `
import * as sc from "./index";
`,
}
