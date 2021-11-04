package gotemplates

var goCommon = map[string]string{
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

var GoTemplates = []map[string]string{
	goCommon,
	constsGo,
	contractGo,
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
