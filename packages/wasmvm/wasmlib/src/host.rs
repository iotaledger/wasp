// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::wasmvmhost::*;

pub trait ScHost {
    fn export_name(&self, index: i32, name: &str);
    fn sandbox(&self, func_nr: i32, params: &[u8]) -> Vec<u8>;
    fn state_delete(&self, key: &[u8]);
    fn state_exists(&self, key: &[u8]) -> bool;
    fn state_get(&self, key: &[u8]) -> Vec<u8>;
    fn state_set(&self, key: &[u8], value: &[u8]);
}

static WASM_VM_HOST: WasmVmHost = WasmVmHost {};
static mut HOST: &dyn ScHost = &WASM_VM_HOST;

pub fn connect_host(h: &'static dyn ScHost) -> &dyn ScHost {
    unsafe {
        let old_host = HOST;
        HOST = h;
        old_host
    }
}

pub fn export_name(index: i32, name: &str) {
    unsafe {
        HOST.export_name(index, name);
    }
}

pub fn sandbox(func_nr: i32, params: &[u8]) -> Vec<u8> {
    unsafe { HOST.sandbox(func_nr, params) }
}

pub fn state_delete(key: &[u8]) {
    unsafe {
        HOST.state_delete(key);
    }
}

pub fn state_exists(key: &[u8]) -> bool {
    unsafe { HOST.state_exists(key) }
}

pub fn state_get(key: &[u8]) -> Vec<u8> {
    unsafe { HOST.state_get(key) }
}

pub fn state_set(key: &[u8], value: &[u8]) {
    unsafe {
        HOST.state_set(key, value);
    }
}
