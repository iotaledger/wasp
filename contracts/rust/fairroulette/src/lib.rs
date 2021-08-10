// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use consts::*;
use fairroulette::*;
use wasmlib::*;

mod consts;
mod fairroulette;
mod types;

#[no_mangle]
fn on_load() {
    let exports = ScExports::new();
    
    exports.add_func(FUNC_PAY_WINNERS, func_pay_winners);
    exports.add_func(FUNC_PLACE_BET, func_place_bet);
    exports.add_func(FUNC_PLAY_PERIOD, func_play_period);
    exports.add_view(VIEW_LAST_WINNING_NUMBER, view_last_winning_number);
}
