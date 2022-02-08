package tstemplates

var stateTs = map[string]string{
	// *******************************
	"state.ts": `
$#emit importWasmTypes
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

export class $TypeName extends wasmtypes.ScProxy {
$#set separator $false
$#if mut stateProxyImmutableFunc
$#each state proxyMethods
}
`,
	// *******************************
	"stateProxyImmutableFunc": `
$#set separator $true
	asImmutable(): sc.Immutable$Package$+State {
		return new sc.Immutable$Package$+State(this.proxy);
	}
`,
}
