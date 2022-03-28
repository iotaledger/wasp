// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var stateGo = map[string]string{
	// *******************************
	"state.go": `
$#emit goPackage

$#emit importWasmTypes
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
	proxy wasmtypes.Proxy
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
