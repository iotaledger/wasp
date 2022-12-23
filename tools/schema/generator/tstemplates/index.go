// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var indexTs = map[string]string{
	// *******************************
	"indexImpl.ts": `
export * from "./funcs";
export * from "./thunks";
`,
	// *******************************
	"index.ts": `
export * from "./consts";
export * from "./contract";
$#set moduleName events
$#if events exportModule
$#set moduleName eventhandlers
$#if events exportModule
$#set moduleName params
$#if params exportModule
$#set moduleName results
$#if results exportModule
$#set moduleName state
$#if state exportModule
$#set moduleName structs
$#if structs exportModule
$#set moduleName typedefs
$#if typedefs exportModule
`,
	// *******************************
	"exportModule": `
export * from "./$moduleName";
`,
}
