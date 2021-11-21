package tstemplates

var indexTs = map[string]string{
	// *******************************
	"index.ts": `
$#if core else exportName
export * from "./consts";
export * from "./contract";
$#if events exportEvents
$#if core else exportKeys
$#if core else exportLib
$#if params exportParams
$#if results exportResults
$#if core else exportRest
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
	"exportKeys": `
export * from "./keys";
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
	"exportRest": `
export * from "./state";
$#if structs exportStructs
$#if typedefs exportTypedefs
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
