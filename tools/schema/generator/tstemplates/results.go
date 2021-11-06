package tstemplates

var resultsTs = map[string]string{
	// *******************************
	"results.ts": `
$#emit importWasmLib
$#emit importSc
$#each func resultsFunc
`,
	// *******************************
	"resultsFunc": `
$#if results resultsFuncResults
`,
	// *******************************
	"resultsFuncResults": `
$#set Kind Result
$#set mut Immutable
$#if result resultsProxyStruct
$#set mut Mutable
$#if result resultsProxyStruct
`,
	// *******************************
	"resultsProxyStruct": `
$#set TypeName $mut$FuncName$+Results
$#each result proxyContainers

export class $TypeName extends wasmlib.ScMapID {
$#each result proxyMethods
}
`,
}
