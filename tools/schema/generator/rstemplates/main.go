// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var mainRs = map[string]string{
	// *******************************
	"main.rs": `
use $package$+impl::*;
use wasmvmhost::*;

#[no_mangle]
fn on_call(index: i32) {
    WasmVmHost::connect();
    on_dispatch(index);
}

#[no_mangle]
fn on_load() {
    WasmVmHost::connect();
    on_dispatch(-1);
}
`,
}
