// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

pub fn func_now(ctx: &ScFuncContext, f: &NowContext) {
    f.state.timestamp().set_value(ctx.timestamp());
}

pub fn view_get_timestamp(_ctx: &ScViewContext, f: &GetTimestampContext) {
    f.results.timestamp().set_value(f.state.timestamp().value());
}
