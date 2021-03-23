// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use consts::*;
use dividend::*;
use wasmlib::*;

mod consts;
mod dividend;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
    exports.add_func(FUNC_DIVIDE, func_divide);
    exports.add_func(FUNC_INIT, func_init);
    exports.add_func(FUNC_MEMBER, func_member);
    exports.add_func(FUNC_SET_OWNER, func_set_owner);
    exports.add_view(VIEW_GET_FACTOR, view_get_factor);
}
