package rstemplates

var resultsRs = map[string]string{
	// *******************************
	"results.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

$#if core useCoreContract useWasmLib
use crate::*;
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
	pub(crate) proxy: Proxy,
}

impl $TypeName {
$#set separator $false
$#each result proxyMethods
}
`,
}
