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
$#if mut stateProxyImmutableFunc
$#each state proxyMethods
`,
	// *******************************
	"stateProxyImmutableFunc": `

func (s $TypeName) AsImmutable() Immutable$Package$+State {
	return Immutable$Package$+State(s)
}
`,
}
