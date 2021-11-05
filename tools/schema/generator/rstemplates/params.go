package rstemplates

var paramsRs = map[string]string{
	// *******************************
	"params.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

use wasmlib::*;
use wasmlib::host::*;
$#if core else paramsUses
$#each func paramsFunc
`,
	// *******************************
	"paramsUses": `

use crate::*;
use crate::keys::*;
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

#[derive(Clone, Copy)]
pub struct $TypeName {
    pub(crate) id: i32,
}

impl $TypeName {
$#each param proxyMethods
}
`,
}

