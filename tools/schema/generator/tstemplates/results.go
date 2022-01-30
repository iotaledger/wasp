package tstemplates

var resultsTs = map[string]string{
	// *******************************
	"results.ts": `
$#emit tsImports
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

export class $TypeName extends wasmtypes.ScProxy {
$#set separator $false
$#each result proxyMethods
}
`,
}
