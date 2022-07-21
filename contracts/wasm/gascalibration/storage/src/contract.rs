// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

#![allow(dead_code)]

use wasmlib::*;
use crate::*;

pub struct FCall {
	pub func: ScFunc,
	pub params: MutableFParams,
}

pub struct ScFuncs {
}

impl ScFuncs {
    pub fn f(_ctx: &dyn ScFuncCallContext) -> FCall {
        let mut f = FCall {
            func: ScFunc::new(HSC_NAME, HFUNC_F),
            params: MutableFParams { proxy: Proxy::nil() },
        };
        ScFunc::link_params(&mut f.params.proxy, &f.func);
        f
    }
}
