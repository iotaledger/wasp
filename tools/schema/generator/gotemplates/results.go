// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var resultsGo = map[string]string{
	// *******************************
	"results.go": `
package $package

$#emit importWasmLibAndWasmTypes
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

type $TypeName struct {
	proxy wasmtypes.Proxy
}
$#if mut resultsMutConstructor
$#each result proxyMethods
`,
	// *******************************
	"resultsMutConstructor": `

func New$TypeName(results *wasmlib.ScDict) $TypeName {
	return $TypeName{proxy: results.AsProxy()}
}
`,
}
