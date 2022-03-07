// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

pub fn func_now(_ctx: &ScFuncContext, _f: &NowContext) {
}

pub fn view_get_timestamp(ctx: &ScViewContext, f: &GetTimestampContext) {
    f.results.timestamp().set_value(ctx.timestamp());
}
