// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var modRs = map[string]string{
	// *******************************
	"mod.rs": `
#![allow(unused_imports)]

pub use consts::*;
pub use contract::*;
$#set moduleName events
$#if events pubUseModule
$#set moduleName eventhandlers
$#if events pubUseModule
$#set moduleName params
$#if params pubUseModule
$#set moduleName results
$#if results pubUseModule
$#set moduleName state
$#if state pubUseModule
$#set moduleName structs
$#if structs pubUseModule
$#set moduleName typedefs
$#if typedefs pubUseModule

pub mod consts;
pub mod contract;
$#set moduleName events
$#if events pubModModule
$#set moduleName eventhandlers
$#if events pubModModule
$#set moduleName params
$#if params pubModModule
$#set moduleName results
$#if results pubModModule
$#set moduleName state
$#if state pubModModule
$#set moduleName structs
$#if structs pubModModule
$#set moduleName typedefs
$#if typedefs pubModModule
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
