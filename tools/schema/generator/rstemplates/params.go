package rstemplates

var paramsRs = map[string]string{
	// *******************************
	"params.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

$#if core useCrate useWasmLib
$#if core useCoreContract
$#if core useHost paramsUses
$#each func paramsFunc
`,
	// *******************************
	"paramsUses": `
use wasmlib::host::*;

use crate::*;
use crate::keys::*;
$#if structs useStructs
$#if typedefs useTypeDefs
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
$#set separator $false
$#each param proxyMethods
}
`,
}
