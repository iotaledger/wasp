// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use consts::*;
use donatewithfeedback::*;
use wasmlib::*;

mod consts;
mod donatewithfeedback;
mod types;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
    exports.add_func(FUNC_DONATE, func_donate);
    exports.add_func(FUNC_WITHDRAW, func_withdraw);
    exports.add_view(VIEW_DONATION, view_donation);
    exports.add_view(VIEW_DONATION_INFO, view_donation_info);
}
