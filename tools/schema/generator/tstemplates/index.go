// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var indexTs = map[string]string{
	// *******************************
	"index.ts": `
$#if core else exportName
export * from "./consts";
export * from "./contract";
$#if events exportEvents
$#if core else exportLib
$#if params exportParams
$#if results exportResults
$#if core else exportState
$#if structs exportStructs
$#if typedefs exportTypedefs
`,
	// *******************************
	"exportName": `
export * from "./$package";

`,
	// *******************************
	"exportEvents": `
export * from "./events";
`,
	// *******************************
	"exportLib": `
export * from "./lib";
`,
	// *******************************
	"exportParams": `
export * from "./params";
`,
	// *******************************
	"exportResults": `
export * from "./results";
`,
	// *******************************
	"exportState": `
export * from "./state";
`,
	// *******************************
	"exportStructs": `
export * from "./structs";
`,
	// *******************************
	"exportTypedefs": `
export * from "./typedefs";
`,
}
