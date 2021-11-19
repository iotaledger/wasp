package rstemplates

var stateRs = map[string]string{
	// *******************************
	"state.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

use wasmlib::*;
use wasmlib::host::*;

use crate::*;
use crate::keys::*;
$#if structs useStructs
$#if typedefs useTypeDefs
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

#[derive(Clone, Copy)]
pub struct $TypeName {
    pub(crate) id: i32,
}
$#if state stateProxyImpl
`,
	// *******************************
	"stateProxyImpl": `

impl $TypeName {
$#set separator $false
$#each state proxyMethods
}
`,
}
