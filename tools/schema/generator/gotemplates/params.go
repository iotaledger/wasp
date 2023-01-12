// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var paramsGo = map[string]string{
	// *******************************
	"params.go": `
package $package

$#emit importWasmLibAndWasmTypes
$#each func paramsFunc
`,
	// *******************************
	"paramsFunc": `
$#if params paramsFuncParams
`,
	// *******************************
	"paramsFuncParams": `
$#set Kind Param
$#set mut Immutable
$#if param paramsProxyStruct
$#set mut Mutable
$#if param paramsProxyStruct
`,
	// *******************************
	"paramsProxyStruct": `
$#set TypeName $mut$FuncName$+Params
$#each param proxyContainers

type $TypeName struct {
	Proxy wasmtypes.Proxy
}
$#if mut else paramsImmutConstructor
$#each param proxyMethods
`,
	// *******************************
	"paramsImmutConstructor": `

func New$TypeName() $TypeName {
	return $TypeName{Proxy: wasmlib.NewParamsProxy()}
}
`,
}
