// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use consts::*;
use helloworld::*;
use wasmlib::*;

mod consts;
mod helloworld;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
    exports.add_func(FUNC_HELLO_WORLD, func_hello_world);
    exports.add_view(VIEW_GET_HELLO_WORLD, view_get_hello_world);
}
