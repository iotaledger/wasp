// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;
use storage::*;
use crate::*;

pub fn func_f(_ctx: &ScFuncContext, f: &FContext) {
    let v = f.state.v();
    let n = f.params.n().value();
    for i in 0..n {
        v.append_uint32().set_value(i);
    }
}
