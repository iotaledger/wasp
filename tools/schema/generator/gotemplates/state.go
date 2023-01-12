// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var stateGo = map[string]string{
	// *******************************
	"state.go": `
package $package

$#emit importWasmLibAndWasmTypes
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
	Proxy wasmtypes.Proxy
}

func New$TypeName() $TypeName {
	return $TypeName{Proxy: wasmlib.NewStateProxy()}
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
