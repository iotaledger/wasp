// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use consts::*;
use inccounter::*;
use wasmlib::*;

mod inccounter;
mod consts;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
    exports.add_func(FUNC_CALL_INCREMENT, func_call_increment);
    exports.add_func(FUNC_CALL_INCREMENT_RECURSE5X, func_call_increment_recurse5x);
    exports.add_func(FUNC_INCREMENT, func_increment);
    exports.add_func(FUNC_INIT, func_init);
    exports.add_func(FUNC_LOCAL_STATE_INTERNAL_CALL, func_local_state_internal_call);
    exports.add_func(FUNC_LOCAL_STATE_POST, func_local_state_post);
    exports.add_func(FUNC_LOCAL_STATE_SANDBOX_CALL, func_local_state_sandbox_call);
    exports.add_func(FUNC_LOOP, func_loop);
    exports.add_func(FUNC_POST_INCREMENT, func_post_increment);
    exports.add_func(FUNC_REPEAT_MANY, func_repeat_many);
    exports.add_func(FUNC_TEST_LEB128, func_test_leb128);
    exports.add_func(FUNC_WHEN_MUST_INCREMENT, func_when_must_increment);
    exports.add_view(VIEW_GET_COUNTER, view_get_counter);
}
