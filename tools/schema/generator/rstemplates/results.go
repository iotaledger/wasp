package rstemplates

var resultsRs = map[string]string{
	// *******************************
	"results.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

use wasmlib::*;
use wasmlib::host::*;
$#if core else resultsUses
$#each func resultsFunc
`,
	// *******************************
	"resultsUses": `

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

