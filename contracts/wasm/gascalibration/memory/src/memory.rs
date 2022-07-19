// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

pub fn func_f(_ctx: &ScFuncContext, f: &FContext) {
    let n = f.params.n().value();
    let mut vec = Vec::<u32>::new();
    for i in 0..n {
        vec.push(i);
    }
}
