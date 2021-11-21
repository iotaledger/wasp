package gotemplates

var GoTemplates = []map[string]string{
	goCommon,
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

var goCommon = map[string]string{
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
}
