package rstemplates

var modRs = map[string]string{
	// *******************************
	"mod.rs": `
#![allow(unused_imports)]

pub use consts::*;
pub use contract::*;
$#set moduleName params
$#if params pubUseModule
$#set moduleName results
$#if results pubUseModule
$#set moduleName structs
$#if structs pubUseModule

pub mod consts;
pub mod contract;
$#set moduleName params
$#if params pubModModule
$#set moduleName results
$#if results pubModModule
$#set moduleName structs
$#if structs pubModModule
`,
	// *******************************
	"pubUseModule": `
pub use $moduleName::*;
`,
	// *******************************
	"pubModModule": `
pub mod $moduleName;
`,
}
