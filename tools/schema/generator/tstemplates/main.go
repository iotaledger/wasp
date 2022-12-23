// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var mainTs = map[string]string{
	// *******************************
	"../main.ts": `
$#emit importWasmVMHost
import * as sc from "./$package$+impl";

export function on_call(index: i32): void {
    wasmvmhost.WasmVMHost.connect();
    sc.onDispatch(index);
}

export function on_load(): void {
    wasmvmhost.WasmVMHost.connect();
    sc.onDispatch(-1);
}
`,
}
