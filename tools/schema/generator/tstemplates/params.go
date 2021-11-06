package tstemplates

var paramsTs = map[string]string{
	// *******************************
	"params.ts": `
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
