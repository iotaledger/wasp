// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;
use solotutorial::*;
use crate::*;

// storeString entry point: stores a string provided as parameters
// in the state as a value of the key 'str'.
// Panics if the parameter is not provided
pub fn func_store_string(_ctx: &ScFuncContext, f: &StoreStringContext) {
    f.state.str().set_value(&f.params.str().value());
}

// getString view: returns the stored string.
pub fn view_get_string(_ctx: &ScViewContext, f: &GetStringContext) {
    f.results.str().set_value(&f.state.str().value());
}
