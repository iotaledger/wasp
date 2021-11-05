package tstemplates

var tsCommon = map[string]string{
	// *******************************
	"tsPackage": `
package $package
`,
	// *******************************
	"importWasmLib": `

import "github.com/iotaledger/wasp/packages/vm/wasmlib/ts/wasmlib"
`,
	// *******************************
	"tsHeader": `
$#emit tsPackage
$#emit importWasmLib
`,
}

var TsTemplates = []map[string]string{
	tsCommon,
	constsTs,
	contractTs,
	funcsTs,
	keysTs,
	libTs,
	mainTs,
	paramsTs,
	proxyTs,
	resultsTs,
	stateTs,
	structsTs,
	typedefsTs,
}
