package gotemplates

var stateGo = map[string]string{
	// *******************************
	"state.go": `
$#emit goPackage
$#if state importWasmLib
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

type $TypeName struct {
	id int32
}
$#each state proxyMethods
`,
}
