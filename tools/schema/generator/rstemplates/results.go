// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var resultsRs = map[string]string{
	// *******************************
	"results.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]
$#if core else useWasmLib
$#emit useCrate
$#if core useCoreContract
$#each func resultsFunc
`,
	// *******************************
	"resultsFunc": `
$#if results resultsFuncResults
`,
	// *******************************
	"resultsFuncResults": `
$#set Kind RESULT_
$#set mut Immutable
$#if result resultsProxyStruct
$#set mut Mutable
$#if result resultsProxyStruct
`,
	// *******************************
	"resultsProxyStruct": `
$#set TypeName $mut$FuncName$+Results
$#each result proxyContainers

#[derive(Clone)]
pub struct $TypeName {
    pub proxy: Proxy,
}

impl $TypeName {
$#set separator $false
$#if mut resultsMutConstructor
$#each result proxyMethods
}
`,
	// *******************************
	"resultsMutConstructor": `
    pub fn new() -> $TypeName {
        $TypeName {
            proxy: results_proxy(),
        }
    }
$#set separator $true
`,
}
