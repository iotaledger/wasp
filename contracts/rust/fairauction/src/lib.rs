// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use consts::*;
use fairauction::*;
use wasmlib::*;

mod consts;
mod fairauction;
mod types;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
    exports.add_func(FUNC_FINALIZE_AUCTION, func_finalize_auction);
    exports.add_func(FUNC_PLACE_BID, func_place_bid);
    exports.add_func(FUNC_SET_OWNER_MARGIN, func_set_owner_margin);
    exports.add_func(FUNC_START_AUCTION, func_start_auction);
    exports.add_view(VIEW_GET_INFO, view_get_info);
}
