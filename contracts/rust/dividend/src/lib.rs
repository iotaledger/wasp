// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use consts::*;
use dividend::*;
use wasmlib::*;

mod consts;
mod dividend;
mod types;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
    exports.add_func(FUNC_DIVIDE, func_divide);
    exports.add_func(FUNC_MEMBER, func_member);
}
