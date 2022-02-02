// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// interface WasmLib to the VM host

// These 2 external functions are funneling the entire
// WasmLib functionality to their counterparts on the host.
#[link(wasm_import_module = "WasmLib")]
extern {
    pub fn hostStateGet(key_ref: *const u8, key_len: i32, val_ref: *const u8, val_len: i32) -> i32;

    pub fn hostStateSet(key_ref: *const u8, key_len: i32, val_ref: *const u8, val_len: i32);
}

pub fn export_name(index: i32, name: &str) {
    unsafe {
        let buf = name.as_bytes();
        hostStateSet(std::ptr::null(), index, buf.as_ptr(), buf.len() as i32);
    }
}

pub fn export_wasm_tag() {
    export_name(-1, "WASM::RUST");
}

pub fn sandbox(func_nr: i32, params: &[u8]) -> Vec<u8> {
    unsafe {
        // call sandbox function, result value will be cached by host
        // always negative funcNr as keyLen indicates sandbox call
        // this removes the need for a separate hostSandbox function
        let size = hostStateGet(std::ptr::null(), func_nr, params.as_ptr(), params.len() as i32);

        // zero length, no need to retrieve cached value
        if size == 0 {
            return Vec::new();
        }

        // retrieve cached value from host
        let mut result = vec![0_u8; size as usize];
        hostStateGet(std::ptr::null(), 0, result.as_mut_ptr(), size);
        result
    }
}

pub fn state_delete(key: &[u8]) {
    unsafe {
        hostStateSet(key.as_ptr(), key.len() as i32, std::ptr::null(), -1);
    }
}

pub fn state_exists(key: &[u8]) -> bool {
    unsafe {
        hostStateGet(key.as_ptr(), key.len() as i32, std::ptr::null(), -1) >= 0
    }
}

pub fn state_get(key: &[u8]) -> Vec<u8> {
    unsafe {
        // variable sized result expected,
        // query size first by passing zero length buffer
        // value will be cached by host
        let size = hostStateGet(key.as_ptr(), key.len() as i32, std::ptr::null(), 0);

        // -1 means non-existent, return default empty buffer
        // zero length, no need to retrieve cached value, return empty buffer
        if size <= 0 {
            return Vec::new();
        }

        // retrieve cached value from host
        let mut result = vec![0_u8; size as usize];
        hostStateGet(std::ptr::null(), 0, result.as_mut_ptr(), size);
        result
    }
}

pub fn state_set(key: &[u8], value: &[u8]) {
    unsafe {
        hostStateSet(key.as_ptr(), key.len() as i32, value.as_ptr(), value.len() as i32);
    }
}
