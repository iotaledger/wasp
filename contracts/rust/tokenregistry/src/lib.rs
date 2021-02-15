// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

#![allow(dead_code)]

use consts::*;
use tokenregistry::*;
use wasmlib::*;

mod consts;
mod tokenregistry;
mod types;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
    exports.add_func(FUNC_MINT_SUPPLY, func_mint_supply);
    exports.add_func(FUNC_TRANSFER_OWNERSHIP, func_transfer_ownership);
    exports.add_func(FUNC_UPDATE_METADATA, func_update_metadata);
    exports.add_view(VIEW_GET_INFO, view_get_info);
}
