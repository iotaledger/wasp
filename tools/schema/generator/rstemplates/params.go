// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var paramsRs = map[string]string{
	// *******************************
	"params.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]
$#if core else useWasmLib
$#emit useCrate
$#if core useCoreContract
$#each func paramsFunc
`,
	// *******************************
	"paramsFunc": `
$#if params paramsFuncParams
`,
	// *******************************
	"paramsFuncParams": `
$#set Kind PARAM_
$#set mut Immutable
$#if param paramsProxyStruct
$#set mut Mutable
$#if param paramsProxyStruct
`,
	// *******************************
	"paramsProxyStruct": `
$#set TypeName $mut$FuncName$+Params
$#each param proxyContainers

#[derive(Clone)]
pub struct $TypeName {
    pub(crate) proxy: Proxy,
}

impl $TypeName {
$#set separator $false
$#each param proxyMethods
}
`,
}
