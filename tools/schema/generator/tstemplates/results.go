package tstemplates

var resultsTs = map[string]string{
	// *******************************
	"results.ts": `
#![allow(dead_code)]
#![allow(unused_imports)]

$#if core useCrate useWasmLib
$#if core useCoreContract
$#if core useHost resultsUses
$#each func resultsFunc
`,
	// *******************************
	"resultsUses": `
use wasmlib::host::*;

use crate::*;
use crate::keys::*;
$#if structs useStructs
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

#[derive(Clone, Copy)]
pub struct $TypeName {
    pub(crate) id: i32,
}

impl $TypeName {
$#each result proxyMethods
}
`,
}
