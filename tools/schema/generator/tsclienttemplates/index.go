// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tsclienttemplates

var indexTs = map[string]string{
	// *******************************
	"index.ts": `
$#if events exportEvents
export * from "./service";
`,
	// *******************************
	"exportEvents": `
export * from "./events";
`,
}
