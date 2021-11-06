package tstemplates

var stateTs = map[string]string{
	// *******************************
	"state.ts": `
$#emit importWasmLib
$#emit importSc
$#set Kind State
$#set mut Immutable
$#emit stateProxyStruct
$#set mut Mutable
$#emit stateProxyStruct
`,
	// *******************************
	"stateProxyStruct": `
$#set TypeName $mut$Package$+State
$#each state proxyContainers

export class $TypeName extends wasmlib.ScMapID {
$#each state proxyMethods
}
`,
}
