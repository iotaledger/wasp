// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var stateRs = map[string]string{
	// *******************************
	"state.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]
$#emit useWasmLib
$#emit useCrate
$#set Kind STATE_
$#set mut Immutable
$#emit stateProxyStruct
$#set mut Mutable
$#emit stateProxyStruct
`,
	// *******************************
	"stateProxyStruct": `
$#set TypeName $mut$Package$+State
$#each state proxyContainers

#[derive(Clone)]
pub struct $TypeName {
    pub(crate) proxy: Proxy,
}
$#if state stateProxyImpl
`,
	// *******************************
	"stateProxyImpl": `

impl $TypeName {
    pub fn new() -> $TypeName {
        $TypeName {
            proxy: state_proxy(),
        }
    }
$#set separator $true
$#if mut stateProxyImmutableFunc
$#each state proxyMethods
}
`,
	// *******************************
	"stateProxyImmutableFunc": `
$#set separator $true
    pub fn as_immutable(&self) -> Immutable$Package$+State {
        Immutable$Package$+State { proxy: self.proxy.root("") }
    }
`,
}
