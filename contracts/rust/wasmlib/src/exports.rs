// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// encapsulates standard host entities into a simple interface

use crate::context::*;
use crate::keys::*;
use crate::mutable::*;

static mut FUNCS: Vec<fn(&ScFuncContext)> = vec![];
static mut VIEWS: Vec<fn(&ScViewContext)> = vec![];

#[no_mangle]
fn on_call_entrypoint(index: i32) {
    unsafe {
        if (index & 0x8000) != 0 {
            VIEWS[(index & 0x7fff) as usize](&ScViewContext {});
            return;
        }

        FUNCS[index as usize](&ScFuncContext {});
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScExports {
    exports: ScMutableStringArray,
}

impl ScExports {
    pub fn new() -> ScExports {
        let exports = ROOT.get_string_array(&KEY_EXPORTS);
        // tell host what our highest predefined key is
        // this helps detect missing or extra keys
        exports.get_string(KEY_ZZZZZZZ.0).set_value("Rust:KEY_ZZZZZZZ");
        ScExports { exports: exports }
    }

    pub fn add_func(&self, name: &str, f: fn(&ScFuncContext)) {
        unsafe {
            let index = FUNCS.len() as i32;
            FUNCS.push(f);
            self.exports.get_string(index).set_value(name);
        }
    }

    pub fn add_view(&self, name: &str, f: fn(&ScViewContext)) {
        unsafe {
            let index = VIEWS.len() as i32;
            VIEWS.push(f);
            self.exports.get_string(index | 0x8000).set_value(name);
        }
    }
}

