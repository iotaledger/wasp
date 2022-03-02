// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var paramsGo = map[string]string{
	// *******************************
	"params.go": `
$#emit goPackage

$#emit importWasmTypes
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
	proxy wasmtypes.Proxy
}
$#each param proxyMethods
`,
}
